// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";

contract PointsToken is ERC20, Ownable {
    constructor(
        uint256 totalSupply
    ) ERC20("Points Token", "PTS") Ownable(msg.sender) {
        _mint(msg.sender, totalSupply * 10 ** decimals());
    }

    // 铸造代币（仅所有者）
    function mint(address to, uint256 amount) external onlyOwner {
        uint256 mintAmount = amount * 10 ** decimals();
        _mint(to, mintAmount);
    }

    // 销毁代币（仅所有者）- 修复小数位处理
    function burn(uint256 amount) external onlyOwner {
        uint256 burnAmount = amount * 10 ** decimals();
        _burn(msg.sender, burnAmount);
    }
}