import { expect } from "chai";
import { ethers, upgrades } from "hardhat";
import { RewardToken, PaymentGateway, Faucet } from "../typechain-types";

describe("RewardToken (UUPS)", function () {
  it("should deploy via proxy with correct name, symbol, and max supply", async function () {
    const [owner] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    expect(await token.name()).to.equal("ShortURL Reward");
    expect(await token.symbol()).to.equal("SURL");
    expect(await token.maxSupply()).to.equal(20_000_000_000n * 10n ** 18n);
  });

  it("should allow owner to mint tokens up to max supply", async function () {
    const [owner] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    await token.mint(owner.address, 1000n * 10n ** 18n);
    expect(await token.balanceOf(owner.address)).to.equal(1000n * 10n ** 18n);
  });

  it("should reject minting beyond max supply", async function () {
    const [owner] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    const max = await token.maxSupply();
    await expect(token.mint(owner.address, max + 1n)).to.be.revertedWithCustomError(token, "ExceedsMaxSupply");
  });

  it("should reject minting from non-owner", async function () {
    const [owner, other] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    await expect(token.connect(other).mint(other.address, 100n)).to.be.revertedWithCustomError(
      token,
      "OwnableUnauthorizedAccount",
    );
  });

  it("should allow token holder to burn tokens", async function () {
    const [owner] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    await token.mint(owner.address, 1000n * 10n ** 18n);
    await token.burn(500n * 10n ** 18n);
    expect(await token.balanceOf(owner.address)).to.equal(500n * 10n ** 18n);
  });

  it("should support upgrades", async function () {
    const [owner] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    const RewardTokenV2Factory = await ethers.getContractFactory("RewardToken");
    await upgrades.upgradeProxy(await token.getAddress(), RewardTokenV2Factory);
    expect(await token.name()).to.equal("ShortURL Reward");
  });
});

describe("PaymentGateway (UUPS)", function () {
  async function deployGatewayFixture() {
    const [owner, user, other] = await ethers.getSigners();
    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    const PaymentGatewayFactory = await ethers.getContractFactory("PaymentGateway");
    const gateway = (await upgrades.deployProxy(PaymentGatewayFactory, [await token.getAddress()], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as PaymentGateway;
    await gateway.waitForDeployment();

    return { token, gateway, owner, user, other };
  }

  it("should accept deposits and emit event", async function () {
    const { token, gateway, user } = await deployGatewayFixture();
    const refId = ethers.keccak256(ethers.toUtf8Bytes("test-ref"));
    const amount = ethers.parseEther("1.0");

    await (token as any).mint(user.address, amount);
    await (token as any).connect(user).approve(await gateway.getAddress(), amount);

    await expect(gateway.connect(user).deposit(refId, amount))
      .to.emit(gateway, "Deposit")
      .withArgs(user.address, refId, amount);

    expect(await gateway.balance()).to.equal(amount);
    expect(await token.balanceOf(await gateway.getAddress())).to.equal(amount);
  });

  it("should reject zero amount deposits", async function () {
    const { gateway, user } = await deployGatewayFixture();
    const refId = ethers.keccak256(ethers.toUtf8Bytes("test-ref"));
    await expect(gateway.connect(user).deposit(refId, 0)).to.be.revertedWithCustomError(
      gateway,
      "InvalidAmount",
    );
  });

  it("should allow owner to withdraw funds", async function () {
    const { token, gateway, owner, user } = await deployGatewayFixture();
    const refId = ethers.keccak256(ethers.toUtf8Bytes("test-ref"));
    const amount = ethers.parseEther("1.0");

    await (token as any).mint(user.address, amount);
    await (token as any).connect(user).approve(await gateway.getAddress(), amount);
    await gateway.connect(user).deposit(refId, amount);

    await expect(gateway.connect(owner).withdraw(user.address, amount))
      .to.changeTokenBalance(token, user, amount);
  });

  it("should reject withdrawal from non-owner", async function () {
    const { gateway, other } = await deployGatewayFixture();
    await expect(gateway.connect(other).withdraw(other.address, 1n)).to.be.revertedWithCustomError(
      gateway,
      "OwnableUnauthorizedAccount",
    );
  });

  it("should support upgrades", async function () {
    const { gateway } = await deployGatewayFixture();
    const V2Factory = await ethers.getContractFactory("PaymentGateway");
    await upgrades.upgradeProxy(await gateway.getAddress(), V2Factory);
    expect(await gateway.balance()).to.equal(0n);
  });
});

describe("Faucet (EIP-712)", function () {
  async function deployFixture() {
    const [owner, signer, user] = await ethers.getSigners();

    const RewardTokenFactory = await ethers.getContractFactory("RewardToken");
    const token = (await upgrades.deployProxy(RewardTokenFactory, ["ShortURL Reward", "SURL", owner.address, ethers.parseUnits("20000000000", 18)], {
      initializer: "initialize",
      kind: "uups",
    })) as unknown as RewardToken;
    await token.waitForDeployment();

    const FaucetFactory = await ethers.getContractFactory("Faucet");
    const faucet = (await FaucetFactory.deploy(
      await token.getAddress(),
      signer.address,
      owner.address,
      ethers.parseUnits("20", 18),
      24 * 60 * 60,
    )) as unknown as Faucet;
    await faucet.waitForDeployment();

    const fundAmount = ethers.parseUnits("10000", 18);
    await token.mint(await faucet.getAddress(), fundAmount);

    return { token, faucet, owner, signer, user, fundAmount };
  }

  function domain(faucetAddress: string, chainId = 31337n) {
    return {
      name: "ShortURL Faucet",
      version: "1",
      chainId,
      verifyingContract: faucetAddress,
    };
  }

  const types = {
    FaucetClaim: [
      { name: "wallet", type: "address" },
      { name: "amount", type: "uint256" },
      { name: "nonce", type: "uint256" },
      { name: "deadline", type: "uint256" },
    ],
  };

  it("should allow claim with valid signature", async function () {
    const { token, faucet, signer, user } = await deployFixture();
    const faucetAddr = await faucet.getAddress();
    const amount = ethers.parseUnits("20", 18);
    const deadline = Math.floor(Date.now() / 1000) + 3600;
    const nonce = 0n;

    const signature = await signer.signTypedData(
      domain(faucetAddr),
      types,
      { wallet: user.address, amount, nonce, deadline },
    );

    await faucet.connect(user).requestTokens(user.address, amount, nonce, deadline, signature);
    expect(await token.balanceOf(user.address)).to.equal(amount);
  });

  it("should reject claim with expired deadline", async function () {
    const { faucet, signer, user } = await deployFixture();
    const faucetAddr = await faucet.getAddress();
    const amount = ethers.parseUnits("20", 18);
    const deadline = Math.floor(Date.now() / 1000) - 60;
    const nonce = 0n;

    const signature = await signer.signTypedData(
      domain(faucetAddr),
      types,
      { wallet: user.address, amount, nonce, deadline },
    );

    await expect(
      faucet.connect(user).requestTokens(user.address, amount, nonce, deadline, signature),
    ).to.be.revertedWithCustomError(faucet, "DeadlineExpired");
  });

  it("should reject claim with invalid signature", async function () {
    const { faucet, signer, user } = await deployFixture();
    const faucetAddr = await faucet.getAddress();
    const amount = ethers.parseUnits("20", 18);
    const deadline = Math.floor(Date.now() / 1000) + 3600;
    const nonce = 0n;

    const payload = { wallet: user.address, amount, nonce, deadline };
    const signature = await signer.signTypedData(domain(faucetAddr), types, payload);

    await expect(
      faucet.connect(user).requestTokens(user.address, amount + 1n, nonce, deadline, signature),
    ).to.be.revertedWithCustomError(faucet, "InvalidSignature");
  });

  it("should reject claim from unauthorized signer", async function () {
    const { faucet, owner, user } = await deployFixture();
    const faucetAddr = await faucet.getAddress();
    const amount = ethers.parseUnits("20", 18);
    const deadline = Math.floor(Date.now() / 1000) + 3600;
    const nonce = 0n;

    const signature = await owner.signTypedData(
      domain(faucetAddr),
      types,
      { wallet: user.address, amount, nonce, deadline },
    );

    await expect(
      faucet.connect(user).requestTokens(user.address, amount, nonce, deadline, signature),
    ).to.be.revertedWithCustomError(faucet, "InvalidSignature");
  });

  it("should enforce cooldown between claims", async function () {
    const { faucet, signer, user } = await deployFixture();
    const faucetAddr = await faucet.getAddress();
    const amount = ethers.parseUnits("10", 18);
    const deadline = Math.floor(Date.now() / 1000) + 3600;
    const nonce = 0n;

    const sig1 = await signer.signTypedData(
      domain(faucetAddr),
      types,
      { wallet: user.address, amount, nonce, deadline },
    );
    await faucet.connect(user).requestTokens(user.address, amount, nonce, deadline, sig1);

    const sig2 = await signer.signTypedData(
      domain(faucetAddr),
      types,
      { wallet: user.address, amount, nonce: 1n, deadline },
    );
    await expect(
      faucet.connect(user).requestTokens(user.address, amount, 1n, deadline, sig2),
    ).to.be.revertedWithCustomError(faucet, "CooldownActive");
  });

  it("should enforce max claim per wallet", async function () {
    const { faucet, signer, user } = await deployFixture();
    const faucetAddr = await faucet.getAddress();
    const amount = ethers.parseUnits("21", 18);
    const deadline = Math.floor(Date.now() / 1000) + 3600;
    const nonce = 0n;

    const signature = await signer.signTypedData(
      domain(faucetAddr),
      types,
      { wallet: user.address, amount, nonce, deadline },
    );

    await expect(
      faucet.connect(user).requestTokens(user.address, amount, nonce, deadline, signature),
    ).to.be.revertedWithCustomError(faucet, "ExceedsMaxClaim");
  });

  it("should allow owner to withdraw tokens", async function () {
    const { token, faucet, owner, fundAmount } = await deployFixture();
    const withdrawAmount = ethers.parseUnits("100", 18);

    await faucet.connect(owner).withdrawTokens(owner.address, withdrawAmount);
    expect(await token.balanceOf(owner.address)).to.equal(withdrawAmount);
    const remainingBal = fundAmount - withdrawAmount;
    expect(await token.balanceOf(await faucet.getAddress())).to.equal(remainingBal);
  });

  it("should allow owner to update signer", async function () {
    const { faucet, owner, signer } = await deployFixture();
    const [newSigner] = await ethers.getSigners();

    await faucet.connect(owner).setSigner(newSigner.address);
    expect(await faucet.authorizedSigner()).to.equal(newSigner.address);
  });
});
