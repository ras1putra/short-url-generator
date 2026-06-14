import { ethers, upgrades, network } from "hardhat";
import * as fs from "fs";
import { TOKEN_CONFIG, ADDRESSES } from "../config";

async function main() {
  const signers = await ethers.getSigners();
  const deployer = signers[0];
  console.log("Deploying Core Contracts with account:", deployer.address);

  const isLocal = network.name === "hardhat" || network.name === "localhost";
  const ownerAddress = isLocal ? signers[1].address : ADDRESSES.owner;
  const faucetSigner = isLocal ? signers[2].address : ADDRESSES.faucetSigner;
  const operatorAddress = isLocal ? signers[2].address : ADDRESSES.operatorSigner;

  const tokenName = TOKEN_CONFIG.name;
  const tokenSymbol = TOKEN_CONFIG.symbol;
  const tokenDecimals = TOKEN_CONFIG.decimals.toString();
  const tokenMaxSupply = TOKEN_CONFIG.maxSupply;

  console.log("Deploying RewardToken proxy...");
  const RewardToken = await ethers.getContractFactory("RewardToken");
  const token = await upgrades.deployProxy(RewardToken, [
    tokenName,
    tokenSymbol,
    deployer.address,
    ethers.parseUnits(tokenMaxSupply, 18)
  ], {
    initializer: "initialize",
    kind: "uups",
  });
  await token.waitForDeployment();
  const tokenAddress = await token.getAddress();

  console.log("Waiting 10 seconds for RPC indexing...");
  await new Promise(resolve => setTimeout(resolve, 10000));

  const tokenImpl = await upgrades.erc1967.getImplementationAddress(tokenAddress);
  console.log("RewardToken proxy:", tokenAddress);
  console.log("RewardToken impl: ", tokenImpl);

  console.log("Deploying PaymentGateway proxy...");
  const PaymentGateway = await ethers.getContractFactory("PaymentGateway");
  const gateway = await upgrades.deployProxy(PaymentGateway, [tokenAddress], {
    initializer: "initialize",
    kind: "uups",
  });
  await gateway.waitForDeployment();
  const gatewayAddress = await gateway.getAddress();

  console.log("Waiting 10 seconds for RPC indexing...");
  await new Promise(resolve => setTimeout(resolve, 10000));

  const gatewayImpl = await upgrades.erc1967.getImplementationAddress(gatewayAddress);
  console.log("PaymentGateway proxy:", gatewayAddress);
  console.log("PaymentGateway impl: ", gatewayImpl);

  console.log("Deploying Faucet...");
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

  console.log("Funding wallets automatically from Deployer...");
  const tokenContract = await ethers.getContractAt("RewardToken", tokenAddress, deployer);

  const faucetFundTx = await tokenContract.mint(faucetAddress, ethers.parseUnits("1000000", 18));
  await faucetFundTx.wait();
  console.log("Faucet funded with 1,000,000 SURL");

  const operatorFundTx = await tokenContract.mint(operatorAddress, ethers.parseUnits("1000000", 18));
  await operatorFundTx.wait();
  console.log("Operator hot wallet funded with 1,000,000 SURL");

  const ownerFundTx = await tokenContract.mint(ownerAddress, ethers.parseUnits("19998000000", 18));
  await ownerFundTx.wait();
  console.log("Owner wallet funded with 19,998,000,000 SURL");

  const transferTokenTx = await tokenContract.transferOwnership(ownerAddress);
  await transferTokenTx.wait();
  console.log(`RewardToken ownership successfully transferred to Owner: ${ownerAddress}`);

  const gatewayContract = await ethers.getContractAt("PaymentGateway", gatewayAddress, deployer);
  const transferGatewayTx = await gatewayContract.transferOwnership(ownerAddress);
  await transferGatewayTx.wait();
  console.log(`PaymentGateway ownership successfully transferred to Owner: ${ownerAddress}`);

  const outputFile = process.env.OUTPUT_FILE || "deployed-addresses.txt";
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
      `FAUCET_SIGNER_PUBLIC_ADDRESS=${faucetSigner}`,
      `OPERATOR_SIGNER_PUBLIC_ADDRESS=${operatorAddress}`,
      "",
    ].join("\n");
    fs.writeFileSync(outputFile, envContent);
    console.log("\nCore addresses written to:", outputFile);
  }

  console.log("\nCore Deployment Summary:");
  console.log("------------------");
  console.log("RewardToken (SURL) proxy: ", tokenAddress);
  console.log("RewardToken implementation:", tokenImpl);
  console.log("PaymentGateway proxy:     ", gatewayAddress);
  console.log("PaymentGateway impl:      ", gatewayImpl);
  console.log("Faucet:                   ", faucetAddress);
  console.log("Owner:                    ", ownerAddress);
}

main().catch((error) => {
  console.error("Core Deployment Failed:", error);
  process.exitCode = 1;
});
