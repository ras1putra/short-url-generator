// SPDX-License-Identifier: MIT
pragma solidity ^0.8.35;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";

contract RewardToken is ERC20Upgradeable, OwnableUpgradeable, UUPSUpgradeable {
    uint256 public maxSupply;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    error ExceedsMaxSupply();
    error InvalidMaxSupply();

    function initialize(
        string memory name,
        string memory symbol,
        address initialOwner,
        uint256 _maxSupply
    ) external initializer {
        if (_maxSupply == 0) revert InvalidMaxSupply();
        __ERC20_init(name, symbol);
        __Ownable_init(initialOwner);
        maxSupply = _maxSupply;
    }

    function mint(address to, uint256 amount) external onlyOwner {
        if (totalSupply() + amount > maxSupply) revert ExceedsMaxSupply();
        _mint(to, amount);
    }

    function burn(uint256 amount) external {
        _burn(msg.sender, amount);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}
}
