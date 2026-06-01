import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@openzeppelin/hardhat-upgrades";
import * as dotenv from "dotenv";
dotenv.config();

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.35",
      settings: {
        optimizer: { enabled: true, runs: 200 },
        evmVersion: "cancun",
      },
  },
  networks: {
    localhost: {
      url: process.env.HARDHAT_RPC_URL || "http://127.0.0.1:8545",
    },
    remote: {
      url: process.env.RPC_URL || "",
      accounts: process.env.DEPLOYER_PRIVATE_KEY ? [process.env.DEPLOYER_PRIVATE_KEY] : [],
    },
  },
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY || "",
  },
};

export default config;
