// SPDX-License-Identifier: MIT
pragma solidity ^0.8.35;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract Faucet is EIP712, ReentrancyGuard, Ownable {
    bytes32 private constant FAUCET_CLAIM_TYPEHASH =
        keccak256("FaucetClaim(address wallet,uint256 amount,uint256 nonce,uint256 deadline)");

    IERC20 public token;
    address public authorizedSigner;

    uint256 public maxClaimPerWallet;
    uint256 public cooldownPeriod;

    mapping(address => uint256) public lastClaimTime;
    mapping(address => uint256) public nonces;

    event ClaimRequested(address indexed wallet, uint256 amount, uint256 nonce);
    event SignerUpdated(address indexed oldSigner, address indexed newSigner);
    event MaxClaimUpdated(uint256 oldMax, uint256 newMax);
    event CooldownUpdated(uint256 oldCooldown, uint256 newCooldown);
    event TokensWithdrawn(address indexed to, uint256 amount);

    error DeadlineExpired();
    error InvalidSignature();
    error CooldownActive(uint256 remaining);
    error ExceedsMaxClaim();
    error InsufficientFaucetBalance();

    constructor(
        address _token,
        address _signer,
        address _owner,
        uint256 _maxClaimPerWallet,
        uint256 _cooldownPeriod
    ) EIP712("ShortURL Faucet", "1") Ownable(_owner) {
        token = IERC20(_token);
        authorizedSigner = _signer;
        maxClaimPerWallet = _maxClaimPerWallet;
        cooldownPeriod = _cooldownPeriod;
    }

    function requestTokens(
        address to,
        uint256 amount,
        uint256 nonce,
        uint256 deadline,
        bytes calldata signature
    ) external nonReentrant {
        if (block.timestamp > deadline) revert DeadlineExpired();

        if (nonces[to] != nonce) revert InvalidSignature();

        bytes32 structHash = keccak256(abi.encode(FAUCET_CLAIM_TYPEHASH, to, amount, nonce, deadline));
        bytes32 digest = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(digest, signature);

        if (signer != authorizedSigner) revert InvalidSignature();

        uint256 lastClaim = lastClaimTime[to];
        if (lastClaim != 0 && block.timestamp < lastClaim + cooldownPeriod) {
            revert CooldownActive(lastClaim + cooldownPeriod - block.timestamp);
        }

        if (amount > maxClaimPerWallet) revert ExceedsMaxClaim();

        if (token.balanceOf(address(this)) < amount) revert InsufficientFaucetBalance();

        nonces[to] = nonce + 1;
        lastClaimTime[to] = block.timestamp;

        require(token.transfer(to, amount), "Transfer failed");

        emit ClaimRequested(to, amount, nonce);
    }

    function setSigner(address newSigner) external onlyOwner {
        if (newSigner == address(0)) revert InvalidSignature();
        emit SignerUpdated(authorizedSigner, newSigner);
        authorizedSigner = newSigner;
    }

    function setMaxClaimPerWallet(uint256 newMax) external onlyOwner {
        emit MaxClaimUpdated(maxClaimPerWallet, newMax);
        maxClaimPerWallet = newMax;
    }

    function setCooldownPeriod(uint256 newCooldown) external onlyOwner {
        emit CooldownUpdated(cooldownPeriod, newCooldown);
        cooldownPeriod = newCooldown;
    }

    function withdrawTokens(address to, uint256 amount) external onlyOwner {
        require(token.transfer(to, amount), "Transfer failed");
        emit TokensWithdrawn(to, amount);
    }

    function faucetBalance() external view returns (uint256) {
        return token.balanceOf(address(this));
    }
}
