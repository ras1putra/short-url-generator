import { ethers, upgrades } from "hardhat";

async function main() {
  const proxyAddress = process.env.PROXY_ADDRESS;
  if (!proxyAddress) {
    throw new Error("PROXY_ADDRESS env var not set");
  }

  console.log("Upgrading proxy at:", proxyAddress);

  const NewContract = await ethers.getContractFactory("RewardToken");
  const upgraded = await upgrades.upgradeProxy(proxyAddress, NewContract);
  await upgraded.waitForDeployment();

  const impl = await upgrades.erc1967.getImplementationAddress(proxyAddress);
  console.log("New implementation:", impl);
  console.log("Upgrade complete!");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
