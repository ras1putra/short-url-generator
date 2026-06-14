export const TOKEN_CONFIG = {
  name: "ShortURL Reward",
  symbol: "SURL",
  decimals: 18,
  maxSupply: "20000000000",
};

export const NFT_PASS_CONFIG = {
  name: "AdBypass NFT Pass",
  symbol: "ABPASS",
  mintPriceSURL: "10000",
  maxSupply: 20000,
  maxMintPerWallet: 1,
  metadataURI: "ipfs://QmProductionVerifiedMetadataHash",
};

export const ADDRESSES = {
  owner: "0x2e78b9d6b3edc8f2fbb6736a4e5b61be8dc7900f",
  faucetSigner: "0x42a17107523adaf9ccab0da4848fac13f36d523b",
  operatorSigner: "0x962a4f31e838ae76c446922da170b41a1e77d3f7",
};

// Production deployed contract addresses (filled manually post-deployment)
export const PRODUCTION_ADDRESSES = {
  token: "",
  payment: "",
  faucet: "",
  nftPass: "",
};
