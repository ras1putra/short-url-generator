// SPDX-License-Identifier: MIT
pragma solidity ^0.8.35;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";

contract RewardToken is ERC20Upgradeable, OwnableUpgradeable, UUPSUpgradeable {
    uint256 public constant MAX_SUPPLY = 20_000_000_000 * 10 ** 18;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(string memory name, string memory symbol, address initialOwner) external initializer {
        __ERC20_init(name, symbol);
        __Ownable_init(initialOwner);
    }

    function mint(address to, uint256 amount) external onlyOwner {
        require(totalSupply() + amount <= MAX_SUPPLY, "Exceeds max supply");
        _mint(to, amount);
    }

    function burn(uint256 amount) external {
        _burn(msg.sender, amount);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}
}
