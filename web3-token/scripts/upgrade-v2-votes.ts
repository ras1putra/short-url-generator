import { ethers, upgrades, network } from "hardhat";

async function main() {
  const isLocal = network.name === "hardhat" || network.name === "localhost";
  let proxyAddress = isLocal
    ? "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"
    : process.env.PROXY_ADDRESS;

  if (!proxyAddress) {
    throw new Error("PROXY_ADDRESS env var not set");
  }
  proxyAddress = proxyAddress.trim();

  console.log("=== RewardToken V1 → V2 Upgrade ===");
  console.log("Proxy address:", proxyAddress);

  const implBefore = await upgrades.erc1967.getImplementationAddress(proxyAddress);
  console.log("Current implementation:", implBefore);

  const signers = await ethers.getSigners();
  const signer = isLocal ? signers[1] : signers[0];

  const TokenV2 = await ethers.getContractFactory("RewardTokenV2", signer);
  const upgraded = await upgrades.upgradeProxy(proxyAddress, TokenV2, {
    kind: "uups",
    call: "initializeV2",
  });
  await upgraded.waitForDeployment();

  const implAfter = await upgrades.erc1967.getImplementationAddress(proxyAddress);
  console.log("New implementation:", implAfter);

  const token = TokenV2.attach(proxyAddress);
  const [name, symbol, maxSupply, totalSupply] = await Promise.all([
    token.name(),
    token.symbol(),
    token.maxSupply(),
    token.totalSupply(),
  ]);

  console.log("\nPost-upgrade state:");
  console.log("  Name:", name);
  console.log("  Symbol:", symbol);
  console.log("  Max supply:", ethers.formatUnits(maxSupply, 18));
  console.log("  Total supply:", ethers.formatUnits(totalSupply, 18));
  console.log("\nVoting features now available:");
  console.log("  delegate(address)");
  console.log("  delegateBySig(address, uint256, uint256, uint8, bytes32, bytes32)");
  console.log("  getVotes(address)");
  console.log("  getPastVotes(address, uint256)");
  console.log("  getPastTotalSupply(uint256)");
  console.log("  delegates(address)");
  console.log("\nUpgrade complete!");
}

main().catch((error) => {
  console.error("Upgrade failed:", error);
  process.exitCode = 1;
});
