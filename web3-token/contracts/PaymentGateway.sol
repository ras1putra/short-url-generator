// SPDX-License-Identifier: MIT
pragma solidity ^0.8.35;

import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract PaymentGateway is OwnableUpgradeable, UUPSUpgradeable {
    event Deposit(address indexed user, bytes32 indexed refId, uint256 amount);

    IERC20 public token;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    error InvalidAmount();
    error TokenTransferFailed();

    function initialize(address _token) external initializer {
        __Ownable_init(msg.sender);
        token = IERC20(_token);
    }

    function deposit(bytes32 refId, uint256 amount) external {
        if (amount == 0) revert InvalidAmount();
        if (!token.transferFrom(msg.sender, address(this), amount)) {
            revert TokenTransferFailed();
        }
        emit Deposit(msg.sender, refId, amount);
    }

    function withdraw(address to, uint256 amount) external onlyOwner {
        if (!token.transfer(to, amount)) {
            revert TokenTransferFailed();
        }
    }

    function balance() external view returns (uint256) {
        return token.balanceOf(address(this));
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}
}
