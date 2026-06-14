// SPDX-License-Identifier: MIT
pragma solidity 0.8.35;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract NFTPass is ERC721, Ownable {
    IERC20 public immutable surlToken;
    uint256 public immutable maxSupply;
    uint256 public mintPrice; // in wei (e.g. 10000 * 10**18)
    uint256 public totalSupply;
    string private _metadataURI;

    error InvalidTokenAddress();
    error InvalidMaxSupply();
    error ExceedsMaxSupply();
    error AlreadyOwnsPass();
    error TokenTransferFailed();

    event Minted(address indexed to, uint256 indexed tokenId);
    event MintPriceUpdated(uint256 newPrice);
    event MetadataURIUpdated(string newURI);

    constructor(
        string memory name,
        string memory symbol,
        address initialOwner,
        address _surlToken,
        uint256 _mintPrice,
        uint256 _maxSupply,
        string memory metadataURI
    ) ERC721(name, symbol) Ownable(initialOwner) {
        if (_surlToken == address(0)) revert InvalidTokenAddress();
        if (_maxSupply == 0) revert InvalidMaxSupply();
        surlToken = IERC20(_surlToken);
        mintPrice = _mintPrice;
        maxSupply = _maxSupply;
        _metadataURI = metadataURI;
    }

    function mint() external {
        if (totalSupply >= maxSupply) revert ExceedsMaxSupply();
        if (balanceOf(msg.sender) > 0) revert AlreadyOwnsPass();

        totalSupply++;
        uint256 tokenId = totalSupply;
        _safeMint(msg.sender, tokenId);

        if (mintPrice > 0) {
            if (!surlToken.transferFrom(msg.sender, owner(), mintPrice)) {
                revert TokenTransferFailed();
            }
        }

        emit Minted(msg.sender, tokenId);
    }

    function setMintPrice(uint256 _mintPrice) external onlyOwner {
        mintPrice = _mintPrice;
        emit MintPriceUpdated(_mintPrice);
    }

    function setMetadataURI(string calldata metadataURI) external onlyOwner {
        _metadataURI = metadataURI;
        emit MetadataURIUpdated(metadataURI);
    }

    function tokenURI(
        uint256 /* tokenId */
    ) public view override returns (string memory) {
        return _metadataURI;
    }
}
