// SPDX-License-Identifier: MIT
pragma solidity ^0.8.35;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract Withdrawer is EIP712, ReentrancyGuard, Ownable {
    bytes32 private constant WITHDRAWAL_TYPEHASH =
        keccak256("Withdrawal(address wallet,uint256 amount,uint256 nonce,uint256 deadline)");

    IERC20 public token;
    address public authorizedSigner;

    mapping(address => uint256) public nonces;

    event WithdrawalClaimed(address indexed wallet, uint256 amount, uint256 nonce);
    event SignerUpdated(address indexed oldSigner, address indexed newSigner);

    error DeadlineExpired();
    error InvalidSignature();
    error InsufficientPoolBalance();
    error InvalidNonce();

    constructor(
        address _token,
        address _signer,
        address _owner
    ) EIP712("ShortURL Withdrawer", "1") Ownable(_owner) {
        token = IERC20(_token);
        authorizedSigner = _signer;
    }

    function withdraw(
        uint256 amount,
        uint256 nonce,
        uint256 deadline,
        bytes calldata signature
    ) external nonReentrant {
        if (block.timestamp > deadline) revert DeadlineExpired();
        if (nonces[msg.sender] != nonce) revert InvalidNonce();

        bytes32 structHash = keccak256(abi.encode(WITHDRAWAL_TYPEHASH, msg.sender, amount, nonce, deadline));
        bytes32 digest = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(digest, signature);
        if (signer != authorizedSigner) revert InvalidSignature();

        if (token.balanceOf(address(this)) < amount) revert InsufficientPoolBalance();

        nonces[msg.sender] = nonce + 1;

        require(token.transfer(msg.sender, amount), "Transfer failed");

        emit WithdrawalClaimed(msg.sender, amount, nonce);
    }

    function depositTokens(uint256 amount) external onlyOwner {
        require(token.transferFrom(msg.sender, address(this), amount), "Transfer failed");
    }

    function poolBalance() external view returns (uint256) {
        return token.balanceOf(address(this));
    }

    function setSigner(address newSigner) external onlyOwner {
        if (newSigner == address(0)) revert InvalidSignature();
        emit SignerUpdated(authorizedSigner, newSigner);
        authorizedSigner = newSigner;
    }

    function emergencyWithdraw(uint256 amount) external onlyOwner {
        require(token.transfer(msg.sender, amount), "Transfer failed");
    }
}
