import { ethers, network } from "hardhat";
import * as fs from "fs";
import * as path from "path";
import { NFT_PASS_CONFIG, ADDRESSES, PRODUCTION_ADDRESSES } from "../config";

async function main() {
  const signers = await ethers.getSigners();
  const deployer = signers[0];
  console.log("Deploying NFTPass with account:", deployer.address);

  const isLocal = network.name === "hardhat" || network.name === "localhost";

  let tokenAddress = "";
  if (isLocal) {
    tokenAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512";
  } else {
    // Read from config
    tokenAddress = PRODUCTION_ADDRESSES.token;
    if (!tokenAddress) {
      throw new Error("RewardToken address not set in PRODUCTION_ADDRESSES.token in config.ts");
    }
  }

  let nftMetadataURI = "";
  if (isLocal) {
    const cidPath = path.join(__dirname, "../metadata-cid.json");
    if (fs.existsSync(cidPath)) {
      const cidData = JSON.parse(fs.readFileSync(cidPath, "utf8"));
      nftMetadataURI = cidData.metadataURI;
      console.log(`Found dynamic metadata CID in dev: ${nftMetadataURI}`);
    } else {
      throw new Error(`Local metadata CID file not found at ${cidPath}. Run upload-assets.ts first!`);
    }
  } else {
    nftMetadataURI = NFT_PASS_CONFIG.metadataURI;
    if (!nftMetadataURI || nftMetadataURI.startsWith("ipfs://QmProductionVerified")) {
      throw new Error("Valid production metadataURI must be configured in config.ts");
    }
  }

  const ownerAddress = isLocal ? signers[1].address : ADDRESSES.owner;

  console.log(`Deploying NFTPass using Token Address: ${tokenAddress}`);
  console.log(`Using Metadata URI: ${nftMetadataURI}`);
  console.log(`Owner address: ${ownerAddress}`);

  const nftName = NFT_PASS_CONFIG.name;
  const nftSymbol = NFT_PASS_CONFIG.symbol;
  const nftMintPrice = NFT_PASS_CONFIG.mintPriceSURL;
  const nftMaxSupply = NFT_PASS_CONFIG.maxSupply;

  const NFTPassFactory = await ethers.getContractFactory("NFTPass");
  const nftPass = await NFTPassFactory.deploy(
    nftName,
    nftSymbol,
    ownerAddress,
    tokenAddress,
    ethers.parseUnits(nftMintPrice, 18),
    nftMaxSupply,
    nftMetadataURI
  );
  await nftPass.waitForDeployment();
  const nftPassAddress = await nftPass.getAddress();
  console.log("NFTPass:                 ", nftPassAddress);

  const outputFile = process.env.OUTPUT_FILE || "deployed-addresses.txt";
  if (outputFile && fs.existsSync(outputFile)) {
    let content = fs.readFileSync(outputFile, "utf8");
    let lines = content.split("\n").map(l => l.trim()).filter(l => l.length > 0);
    lines = lines.filter(line => !line.startsWith("CONTRACT_NFT_PASS="));
    lines.push(`CONTRACT_NFT_PASS=${nftPassAddress}`);
    fs.writeFileSync(outputFile, lines.join("\n") + "\n");
    console.log("Updated addresses output file with CONTRACT_NFT_PASS.");
  } else if (outputFile) {
    const envContent = `CONTRACT_NFT_PASS=${nftPassAddress}\n`;
    fs.writeFileSync(outputFile, envContent);
    console.log("Created address output file with CONTRACT_NFT_PASS.");
  }
}

main().catch((error) => {
  console.error("NFT Pass Deployment Failed:", error);
  process.exitCode = 1;
});
