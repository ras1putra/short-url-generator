import { expect } from "chai";
import { ethers, upgrades } from "hardhat";
import { mine } from "@nomicfoundation/hardhat-network-helpers";
import { RewardToken, RewardTokenV2 } from "../typechain-types";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";

describe("RewardTokenV2 (Upgrade + Voting)", function () {
  let tokenV1: RewardToken;
  let tokenV2: RewardTokenV2;
  let proxyAddress: string;
  let owner: HardhatEthersSigner;
  let alice: HardhatEthersSigner;
  let bob: HardhatEthersSigner;

  const NAME = "ShortURL Reward";
  const SYMBOL = "SURL";
  const MAX_SUPPLY = ethers.parseUnits("20000000000", 18);
  const MINT_AMOUNT = ethers.parseUnits("1000000", 18);

  async function deployV1() {
    const signers = await ethers.getSigners();
    owner = signers[0];
    alice = signers[1];
    bob = signers[2];

    const V1Factory = await ethers.getContractFactory("RewardToken");
    tokenV1 = (await upgrades.deployProxy(
      V1Factory,
      [NAME, SYMBOL, owner.address, MAX_SUPPLY],
      { initializer: "initialize", kind: "uups" },
    )) as unknown as RewardToken;
    await tokenV1.waitForDeployment();
    proxyAddress = await tokenV1.getAddress();
  }

  async function upgradeToV2() {
    const V2Factory = await ethers.getContractFactory("RewardTokenV2");
    const upgraded = await upgrades.upgradeProxy(proxyAddress, V2Factory, {
      kind: "uups",
      call: "initializeV2",
    });
    await upgraded.waitForDeployment();
    tokenV2 = V2Factory.attach(proxyAddress) as unknown as RewardTokenV2;
  }

  describe("Upgrade Safety", function () {
    it("should preserve storage after upgrade", async function () {
      await deployV1();
      await tokenV1.mint(alice.address, MINT_AMOUNT);
      await tokenV1.mint(bob.address, MINT_AMOUNT);

      await upgradeToV2();

      expect(await tokenV2.name()).to.equal(NAME);
      expect(await tokenV2.symbol()).to.equal(SYMBOL);
      expect(await tokenV2.maxSupply()).to.equal(MAX_SUPPLY);
      expect(await tokenV2.balanceOf(alice.address)).to.equal(MINT_AMOUNT);
      expect(await tokenV2.balanceOf(bob.address)).to.equal(MINT_AMOUNT);
      expect(await tokenV2.totalSupply()).to.equal(MINT_AMOUNT * 2n);
      expect(await tokenV2.owner()).to.equal(owner.address);
    });

    it("should keep same proxy address after upgrade", async function () {
      await deployV1();
      const addrBefore = proxyAddress;
      await upgradeToV2();
      expect(await tokenV2.getAddress()).to.equal(addrBefore);
    });

    it("should reject upgrade from non-owner", async function () {
      await deployV1();
      const V2Factory = await ethers.getContractFactory("RewardTokenV2", alice);
      await expect(
        upgrades.upgradeProxy(proxyAddress, V2Factory, { kind: "uups" }),
      ).to.be.revertedWithCustomError(tokenV1, "OwnableUnauthorizedAccount");
    });

    it("should allow mint/burn after upgrade", async function () {
      await deployV1();
      await upgradeToV2();

      await tokenV2.mint(alice.address, MINT_AMOUNT);
      expect(await tokenV2.balanceOf(alice.address)).to.equal(MINT_AMOUNT);

      await tokenV2.connect(alice).burn(MINT_AMOUNT / 2n);
      expect(await tokenV2.balanceOf(alice.address)).to.equal(MINT_AMOUNT / 2n);
    });

    it("should enforce maxSupply after upgrade", async function () {
      await deployV1();
      await upgradeToV2();

      await expect(
        tokenV2.mint(alice.address, MAX_SUPPLY + 1n),
      ).to.be.revertedWithCustomError(tokenV2, "ExceedsMaxSupply");
    });
  });

  describe("Delegation", function () {
    beforeEach(async function () {
      await deployV1();
      await tokenV1.mint(alice.address, MINT_AMOUNT);
      await tokenV1.mint(bob.address, MINT_AMOUNT);
      await upgradeToV2();
    });

    it("should start with no delegation (votes = 0)", async function () {
      expect(await tokenV2.getVotes(alice.address)).to.equal(0n);
      expect(await tokenV2.delegates(alice.address)).to.equal(ethers.ZeroAddress);
    });

    it("should delegate to self", async function () {
      await tokenV2.connect(alice).delegate(alice.address);

      expect(await tokenV2.delegates(alice.address)).to.equal(alice.address);
      expect(await tokenV2.getVotes(alice.address)).to.equal(MINT_AMOUNT);
    });

    it("should delegate to another address", async function () {
      await tokenV2.connect(alice).delegate(bob.address);

      expect(await tokenV2.delegates(alice.address)).to.equal(bob.address);
      expect(await tokenV2.getVotes(bob.address)).to.equal(MINT_AMOUNT);
      expect(await tokenV2.getVotes(alice.address)).to.equal(0n);
    });

    it("should accumulate delegated votes from multiple delegators", async function () {
      await tokenV2.connect(alice).delegate(bob.address);
      await tokenV2.connect(bob).delegate(bob.address);

      expect(await tokenV2.getVotes(bob.address)).to.equal(MINT_AMOUNT * 2n);
    });

    it("should update votes when balance changes after delegation", async function () {
      await tokenV2.connect(alice).delegate(alice.address);
      expect(await tokenV2.getVotes(alice.address)).to.equal(MINT_AMOUNT);

      const transferAmount = ethers.parseUnits("100000", 18);
      await tokenV2.connect(alice).transfer(bob.address, transferAmount);

      expect(await tokenV2.getVotes(alice.address)).to.equal(MINT_AMOUNT - transferAmount);
    });

    it("should update votes when tokens are minted to delegated address", async function () {
      await tokenV2.connect(alice).delegate(alice.address);
      const votesBefore = await tokenV2.getVotes(alice.address);

      const extra = ethers.parseUnits("500000", 18);
      await tokenV2.mint(alice.address, extra);

      expect(await tokenV2.getVotes(alice.address)).to.equal(votesBefore + extra);
    });

    it("should update votes when tokens are burned from delegated address", async function () {
      const mintAmount = ethers.parseUnits("500000", 18);
      await tokenV2.mint(alice.address, mintAmount);

      await tokenV2.connect(alice).delegate(alice.address);
      const votesBefore = await tokenV2.getVotes(alice.address);

      const burnAmount = ethers.parseUnits("200000", 18);
      await tokenV2.connect(alice).burn(burnAmount);

      expect(await tokenV2.getVotes(alice.address)).to.equal(votesBefore - burnAmount);
    });

    it("should handle re-delegation", async function () {
      await tokenV2.connect(alice).delegate(alice.address);
      expect(await tokenV2.getVotes(alice.address)).to.equal(MINT_AMOUNT);

      await tokenV2.connect(alice).delegate(bob.address);
      expect(await tokenV2.getVotes(alice.address)).to.equal(0n);
      expect(await tokenV2.getVotes(bob.address)).to.equal(MINT_AMOUNT);
    });

    it("should emit DelegateChanged event", async function () {
      await expect(tokenV2.connect(alice).delegate(bob.address))
        .to.emit(tokenV2, "DelegateChanged")
        .withArgs(alice.address, ethers.ZeroAddress, bob.address);
    });

    it("should emit DelegateVotesChanged event", async function () {
      await expect(tokenV2.connect(alice).delegate(alice.address))
        .to.emit(tokenV2, "DelegateVotesChanged")
        .withArgs(alice.address, 0n, MINT_AMOUNT);
    });
  });

  describe("Checkpoints & Past Votes", function () {
    beforeEach(async function () {
      await deployV1();
      await tokenV1.mint(alice.address, MINT_AMOUNT);
      await upgradeToV2();
    });

    it("should track checkpoints on delegation", async function () {
      await tokenV2.connect(alice).delegate(alice.address);
      expect(await tokenV2.numCheckpoints(alice.address)).to.equal(1n);
    });

    it("should track multiple checkpoints on balance changes", async function () {
      await tokenV2.connect(alice).delegate(alice.address);
      const cp0 = await tokenV2.checkpoints(alice.address, 0);

      await tokenV2.connect(alice).transfer(bob.address, ethers.parseUnits("100000", 18));
      const cp1 = await tokenV2.checkpoints(alice.address, 1);

      expect(cp0._value).to.equal(MINT_AMOUNT);
      expect(cp1._value).to.equal(MINT_AMOUNT - ethers.parseUnits("100000", 18));
    });

    it("should return correct getPastVotes", async function () {
      const delegateTx = await tokenV2.connect(alice).delegate(alice.address);
      const delegateReceipt = await delegateTx.wait();
      const delegateBlock = delegateReceipt!.blockNumber;

      const transferTx = await tokenV2.connect(alice).transfer(bob.address, ethers.parseUnits("100000", 18));
      const transferReceipt = await transferTx.wait();
      const transferBlock = transferReceipt!.blockNumber;

      await mine();

      expect(await tokenV2.getPastVotes(alice.address, delegateBlock)).to.equal(MINT_AMOUNT);
      expect(await tokenV2.getPastVotes(alice.address, transferBlock)).to.equal(
        MINT_AMOUNT - ethers.parseUnits("100000", 18),
      );
    });

    it("should revert getPastVotes for current/future block", async function () {
      const currentBlock = await ethers.provider.getBlockNumber();
      await expect(
        tokenV2.getPastVotes(alice.address, currentBlock),
      ).to.be.revertedWithCustomError(tokenV2, "ERC5805FutureLookup");
    });

    it("should return correct getPastTotalSupply", async function () {
      const mintAmount = ethers.parseUnits("500000", 18);
      const mintTx = await tokenV2.mint(alice.address, mintAmount);
      const mintReceipt = await mintTx.wait();
      const mintBlock = mintReceipt!.blockNumber;

      await tokenV2.connect(alice).delegate(alice.address);
      const burnAmount = ethers.parseUnits("100000", 18);
      const burnTx = await tokenV2.connect(alice).burn(burnAmount);
      const burnReceipt = await burnTx.wait();
      const burnBlock = burnReceipt!.blockNumber;

      await mine();

      expect(await tokenV2.getPastTotalSupply(mintBlock)).to.equal(mintAmount);
      expect(await tokenV2.getPastTotalSupply(burnBlock)).to.equal(mintAmount - burnAmount);
    });
  });

  describe("delegateBySig", function () {
    beforeEach(async function () {
      await deployV1();
      await tokenV1.mint(alice.address, MINT_AMOUNT);
      await upgradeToV2();
    });

    it("should delegate via EIP-712 signature", async function () {
      const chainId = (await ethers.provider.getNetwork()).chainId;
      const domain = {
        name: NAME,
        version: "1",
        chainId,
        verifyingContract: proxyAddress,
      };

      const types = {
        Delegation: [
          { name: "delegatee", type: "address" },
          { name: "nonce", type: "uint256" },
          { name: "expiry", type: "uint256" },
        ],
      };

      const expiry = Math.floor(Date.now() / 1000) + 3600;
      const nonce = 0n;

      const signature = await alice.signTypedData(domain, types, {
        delegatee: bob.address,
        nonce,
        expiry,
      });
      const { v, r, s } = ethers.Signature.from(signature);

      await tokenV2.delegateBySig(bob.address, nonce, expiry, v, r, s);

      expect(await tokenV2.delegates(alice.address)).to.equal(bob.address);
      expect(await tokenV2.getVotes(bob.address)).to.equal(MINT_AMOUNT);
    });

    it("should reject expired signature", async function () {
      const chainId = (await ethers.provider.getNetwork()).chainId;
      const domain = {
        name: NAME,
        version: "1",
        chainId,
        verifyingContract: proxyAddress,
      };

      const types = {
        Delegation: [
          { name: "delegatee", type: "address" },
          { name: "nonce", type: "uint256" },
          { name: "expiry", type: "uint256" },
        ],
      };

      const expiry = Math.floor(Date.now() / 1000) - 60;
      const nonce = 0n;

      const signature = await alice.signTypedData(domain, types, {
        delegatee: bob.address,
        nonce,
        expiry,
      });
      const { v, r, s } = ethers.Signature.from(signature);

      await expect(
        tokenV2.delegateBySig(bob.address, nonce, expiry, v, r, s),
      ).to.be.revertedWithCustomError(tokenV2, "VotesExpiredSignature");
    });

    it("should reject reused nonce", async function () {
      const chainId = (await ethers.provider.getNetwork()).chainId;
      const domain = {
        name: NAME,
        version: "1",
        chainId,
        verifyingContract: proxyAddress,
      };

      const types = {
        Delegation: [
          { name: "delegatee", type: "address" },
          { name: "nonce", type: "uint256" },
          { name: "expiry", type: "uint256" },
        ],
      };

      const expiry = Math.floor(Date.now() / 1000) + 3600;

      const sig1 = await alice.signTypedData(domain, types, {
        delegatee: bob.address,
        nonce: 0n,
        expiry,
      });
      const { v: v1, r: r1, s: s1 } = ethers.Signature.from(sig1);
      await tokenV2.delegateBySig(bob.address, 0n, expiry, v1, r1, s1);

      const sig2 = await alice.signTypedData(domain, types, {
        delegatee: bob.address,
        nonce: 0n,
        expiry,
      });
      const { v: v2, r: r2, s: s2 } = ethers.Signature.from(sig2);
      await expect(
        tokenV2.delegateBySig(bob.address, 0n, expiry, v2, r2, s2),
      ).to.be.revertedWithCustomError(tokenV2, "InvalidAccountNonce");
    });
  });

  describe("ERC20Votes Upgrade Compatibility", function () {
    it("should deploy V2 directly (fresh, no upgrade)", async function () {
      const signers = await ethers.getSigners();
      const deployer = signers[0];

      const V2Factory = await ethers.getContractFactory("RewardTokenV2");
      const freshToken = (await upgrades.deployProxy(
        V2Factory,
        [NAME, SYMBOL, deployer.address, MAX_SUPPLY],
        { initializer: "initialize", kind: "uups" },
      )) as unknown as RewardTokenV2;
      await freshToken.waitForDeployment();

      expect(await freshToken.name()).to.equal(NAME);
      expect(await freshToken.symbol()).to.equal(SYMBOL);
      expect(await freshToken.maxSupply()).to.equal(MAX_SUPPLY);
    });

    it("should handle upgrade with existing delegation (V1 had no votes)", async function () {
      await deployV1();
      await tokenV1.mint(alice.address, MINT_AMOUNT);
      await tokenV1.connect(alice).transfer(bob.address, ethers.parseUnits("500000", 18));

      await upgradeToV2();

      expect(await tokenV2.getVotes(alice.address)).to.equal(0n);
      expect(await tokenV2.getVotes(bob.address)).to.equal(0n);

      await tokenV2.connect(alice).delegate(alice.address);
      expect(await tokenV2.getVotes(alice.address)).to.equal(ethers.parseUnits("500000", 18));
    });
  });
});
