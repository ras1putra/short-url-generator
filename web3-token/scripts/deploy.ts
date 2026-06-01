import { ethers, upgrades, network } from "hardhat";
import * as fs from "fs";

async function main() {
  const signers = await ethers.getSigners();
  const deployer = signers[0];
  console.log("Deploying with account:", deployer.address);

  const isDevMode = process.env.APP_MODE === "dev" || network.name === "hardhat" || network.name === "localhost";

  let ownerAddress = process.env.OWNER_ADDRESS;
  if (isDevMode) {
    ownerAddress = signers[1].address;
  } else if (!ownerAddress) {
    throw new Error("OWNER_ADDRESS environment variable is required for non-development networks");
  }

  const tokenName = process.env.TOKEN_NAME || "ShortURL Reward";
  const tokenSymbol = process.env.TOKEN_SYMBOL || "SURL";
  const tokenDecimals = process.env.TOKEN_DECIMALS || "18";
  const RewardToken = await ethers.getContractFactory("RewardToken");
  const token = await upgrades.deployProxy(RewardToken, [tokenName, tokenSymbol, ownerAddress], {
    initializer: "initialize",
    kind: "uups",
  });
  await token.waitForDeployment();
  const tokenAddress = await token.getAddress();

  const tokenImpl = await upgrades.erc1967.getImplementationAddress(tokenAddress);
  console.log("RewardToken proxy:", tokenAddress);
  console.log("RewardToken impl: ", tokenImpl);

  const PaymentGateway = await ethers.getContractFactory("PaymentGateway");
  const gateway = await upgrades.deployProxy(PaymentGateway, [tokenAddress], {
    initializer: "initialize",
    kind: "uups",
  });
  await gateway.waitForDeployment();
  const gatewayAddress = await gateway.getAddress();

  const gatewayImpl = await upgrades.erc1967.getImplementationAddress(gatewayAddress);
  console.log("PaymentGateway proxy:", gatewayAddress);
  console.log("PaymentGateway impl: ", gatewayImpl);


  let faucetSigner = process.env.FAUCET_SIGNER_ADDRESS;
  if (isDevMode) {
    faucetSigner = signers[2].address;
  } else if (!faucetSigner) {
    throw new Error("FAUCET_SIGNER_ADDRESS environment variable is required for non-development networks");
  }

  const FaucetFactory = await ethers.getContractFactory("Faucet");
  const faucet = await FaucetFactory.deploy(
    tokenAddress,
    faucetSigner,
    ownerAddress,
    ethers.parseUnits("20", 18),
    24 * 60 * 60, // 24 hours cooldown
  );
  await faucet.waitForDeployment();
  const faucetAddress = await faucet.getAddress();
  console.log("Faucet:                  ", faucetAddress);

  let operatorAddress = process.env.OPERATOR_SIGNER_ADDRESS;
  if (isDevMode) {
    operatorAddress = faucetSigner;
  } else if (!operatorAddress) {
    throw new Error("OPERATOR_SIGNER_ADDRESS environment variable is required for non-development networks");
  }

  if (isDevMode) {
    const ownerSigner = signers[1];
    const tokenContract = await ethers.getContractAt("RewardToken", tokenAddress, ownerSigner);
    const fundAmount = ethers.parseUnits("1000000", 18); // 1M SURL
    const faucetFundTx = await tokenContract.mint(faucetAddress, fundAmount);
    await faucetFundTx.wait();
    console.log("Faucet funded with 1,000,000 SURL");
    const operatorFundTx = await tokenContract.mint(operatorAddress, fundAmount);
    await operatorFundTx.wait();
    console.log("Operator hot wallet funded with 1,000,000 SURL");
  } else {
    console.log(`Please fund the faucet manually. Faucet address: ${faucetAddress}`);
    console.log(`Please fund the operator hot wallet manually (ETH + SURL). Operator address: ${operatorAddress}`);
  }

  const outputFile = process.env.OUTPUT_FILE;
  if (outputFile) {
    const envContent = [
      `CONTRACT_TOKEN=${tokenAddress}`,
      `CONTRACT_PAYMENT=${gatewayAddress}`,
      `CONTRACT_FAUCET=${faucetAddress}`,
      `CONTRACT_TOKEN_IMPL=${tokenImpl}`,
      `CONTRACT_PAYMENT_IMPL=${gatewayImpl}`,
      `OWNER_ADDRESS=${ownerAddress}`,
      `DEPLOYER_ADDRESS=${deployer.address}`,
      `TOKEN_SYMBOL=${tokenSymbol}`,
      `TOKEN_DECIMALS=${tokenDecimals}`,
      `FAUCET_SIGNER_ADDRESS=${faucetSigner}`,
      `OPERATOR_SIGNER_ADDRESS=${operatorAddress}`,
      "",
    ].join("\n");
    fs.writeFileSync(outputFile, envContent);
    console.log("\nAddresses written to:", outputFile);
  }

  console.log("\nDeployment Summary:");
  console.log("------------------");
  console.log("RewardToken (SURL) proxy: ", tokenAddress);
  console.log("RewardToken implementation:", tokenImpl);
  console.log("PaymentGateway proxy:     ", gatewayAddress);
  console.log("PaymentGateway impl:      ", gatewayImpl);
  console.log("Faucet:                   ", faucetAddress);
  console.log("Owner:                    ", ownerAddress);
  console.log("Faucet Signer:            ", faucetSigner);
  console.log("Operator:                 ", operatorAddress);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
