package service

import (
	"POINTSTOKEN/config"
	"POINTSTOKEN/db"
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

const (
	LogMakeTopic   = "0xfc37f2ff950f95913eb7182357ba3c14df60ef354bc7d6ab1ba2815f249fffe6"
	LogCancelTopic = "0x0ac8bb53fac566d7afc05d8b4df11d7690a7b27bdc40b54e4060f9b21fb849bd"
	LogMatchTopic  = "0xf629aecab94607bc43ce4aebd564bf6e61c7327226a797b002de724b9944b20e"
	contractAbi    = `[
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "totalSupply",
				"type": "uint256"
			}
		],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "allowance",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "needed",
				"type": "uint256"
			}
		],
		"name": "ERC20InsufficientAllowance",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "sender",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "balance",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "needed",
				"type": "uint256"
			}
		],
		"name": "ERC20InsufficientBalance",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "approver",
				"type": "address"
			}
		],
		"name": "ERC20InvalidApprover",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "receiver",
				"type": "address"
			}
		],
		"name": "ERC20InvalidReceiver",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "sender",
				"type": "address"
			}
		],
		"name": "ERC20InvalidSender",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			}
		],
		"name": "ERC20InvalidSpender",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "owner",
				"type": "address"
			}
		],
		"name": "OwnableInvalidOwner",
		"type": "error"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "account",
				"type": "address"
			}
		],
		"name": "OwnableUnauthorizedAccount",
		"type": "error"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "owner",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "spender",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Approval",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "previousOwner",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "OwnershipTransferred",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "owner",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			}
		],
		"name": "allowance",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "spender",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "approve",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "account",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "burn",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "decimals",
		"outputs": [
			{
				"internalType": "uint8",
				"name": "",
				"type": "uint8"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "mint",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "name",
		"outputs": [
			{
				"internalType": "string",
				"name": "",
				"type": "string"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "owner",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "renounceOwnership",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "symbol",
		"outputs": [
			{
				"internalType": "string",
				"name": "",
				"type": "string"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalSupply",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "to",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "transferFrom",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "transferOwnership",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
)

// ChainManager 链管理器，管理多链连接和时间监听
type ChainManager struct {
	chains     map[uint64]*ChainHandler
	repository db.Repository
	wg         sync.WaitGroup
}

// ChainHandler 单链处理器
type ChainHandler struct {
	config        config.ChainConfig
	client        *ethclient.Client
	contract      common.Address
	repository    db.Repository
	confirmations uint64
}

// NewChainManager 创建新的链管理器
func NewChainManager(chainsConfig []config.ChainConfig, repo db.Repository) (*ChainManager, error) {
	manager := &ChainManager{
		chains:     make(map[uint64]*ChainHandler),
		repository: repo,
	}

	for _, chainConfig := range chainsConfig {
		handler, err := NewChainHandler(chainConfig, repo)
		if err != nil {
			return nil, err
		}

		manager.chains[chainConfig.ChainID] = handler
	}
	return manager, nil
}

func NewChainHandler(cfg config.ChainConfig, repository db.Repository) (*ChainHandler, error) {
	//连接到链
	client, err := ethclient.Dial(cfg.RPCUrl)
	if err != nil {
		return nil, err
	}

	//验证链id
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	if chainID.Uint64() != cfg.ChainID {
		return nil, fmt.Errorf("chainID %d is not equal to %d", chainID.Uint64(), cfg.ChainID)
	}

	return &ChainHandler{
		config:        cfg,
		client:        client,
		contract:      common.HexToAddress(cfg.ContractAddr),
		repository:    repository,
		confirmations: cfg.Confirmations,
	}, nil
}

// StartEventListeners 启动所有链的事件监听
func (m *ChainManager) StartEventListeners(ctx context.Context) {
	for _, handler := range m.chains {
		m.wg.Add(1)
		go func(h *ChainHandler) {
			defer m.wg.Done()
			h.StartEVentListener(ctx)
		}(handler)
	}
}

// StartEVentListener 启动事件监控
func (h *ChainHandler) StartEVentListener(ctx context.Context) {
	log.Printf("Starting EVent listener for chain %s (ID : %d)", h.config.Name, h.config.ChainID)

	//获取最后处理的区快
	lastBlock, err := h.repository.GetChainLastBlock(h.config.ChainID)
	if err != nil {
		lastBlock = h.config.StartBlock
	}

	if lastBlock == 0 {
		header, err := h.client.HeaderByNumber(ctx, nil)
		if err != nil {
			log.Printf("failed to fetch header for chain %s, %s", h.config.Name, h.config.ChainID)
			return
		}
		lastBlock = header.Number.Uint64() - 100
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping EVent listener for chain %s", h.config.Name)
			return
		default:
			latestHeader, err := h.client.HeaderByNumber(ctx, nil)
			if err != nil {
				log.Printf("failed to get latest block header  %v", h.config.Name)
				time.Sleep(h.config.PollInterval)
				continue
			}
			latestBlock := latestHeader.Number.Uint64()

			//只处理到确认后的区快
			if latestBlock <= h.confirmations {
				time.Sleep(h.config.PollInterval)
				continue
			}

			safeBlock := latestBlock - h.confirmations
			// 如果没有新的区块需要处理
			if lastBlock >= safeBlock {
				time.Sleep(h.config.PollInterval)
				continue
			}
			// 处理从lastBlock+1到safeBlock的区块
			log.Printf("Processing blocks %d to %d on chain %s", lastBlock+1, safeBlock, h.config.Name)
			if err := h.processBlocks(ctx, lastBlock+1, safeBlock); err != nil {
				log.Printf("Error processing blocks on chain %s: %v", h.config.Name, err)
				time.Sleep(h.config.PollInterval)
				continue
			}

			// 更新最后处理的区块
			lastBlock = safeBlock
			if err := h.repository.UpdateChainLastBlock(h.config.ChainID, lastBlock); err != nil {
				log.Printf("Failed to update last processed block: %v", err)
			}
		}
	}
}

// 处理指定范围内的区快
func (h *ChainHandler) processBlocks(ctx context.Context, start uint64, end uint64) error {
	reader := bytes.NewReader([]byte(contractAbi))
	parsedABI, err := abi.JSON(reader)
	if err != nil {
		log.Fatalf("解析ABI失败: %v", err)
		return err
	}

	contractAddr := common.HexToAddress(h.config.ContractAddr)

	startBlock := new(big.Int).SetUint64(start)
	endBlock := new(big.Int).SetUint64(end)
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{contractAddr},
		Topics: [][]common.Hash{
			{parsedABI.Events["Transfer"].ID}, // 过滤Transfer事件
		},
	}
	logs, err := h.client.FilterLogs(ctx, query)
	if err != nil {
		log.Fatalf("查询日志失败: %v", err)
		//return err
	}

	fmt.Printf("找到 %d 笔相关交易:\n", len(logs))
	for _, vLog := range logs {
		// 解析事件数据
		var transferEvent struct {
			From  common.Address
			To    common.Address
			Value *big.Int
		}
		if err := parsedABI.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data); err != nil {
			log.Printf("解析事件失败: %v", err)
			continue
		}
		// 从日志中获取索引字段
		transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
		transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
		evenType := "transfer"
		zeroAddr := common.Address{}.Hex()
		if transferEvent.From.Hex() == zeroAddr {
			evenType = "mint"
		} else if transferEvent.To.Hex() == zeroAddr {
			evenType = "burn"
		}

		// 获取用户token 余额
		balanceFrom, _, _, err := getUserTokenBalance(context.Background(), h.client, h.contract, transferEvent.From)
		if err != nil {
			return err
		}
		balanceTo, _, _, err := getUserTokenBalance(context.Background(), h.client, h.contract, transferEvent.To)
		if err != nil {
			return err
		}
		err = h.repository.UpdateUserBalance(h.config.ChainID, transferEvent.From, balanceFrom.Uint64())
		if err != nil {
			return err
		}
		log.Printf("区块BlockNumber: %v ", vLog.BlockNumber)

		block, err := h.client.BlockByNumber(context.Background(), new(big.Int).SetUint64(vLog.BlockNumber))
		if err != nil {
			log.Printf("获取区块失败（高度: %d）: %v", vLog.BlockNumber, err)
			continue
		}
		timeStamp := time.Unix(int64(block.Time()), 0)
		//timeStamp := time.Unix(int64(vLog.BlockTimestamp), 0).In(time.FixedZone("CST", 8*3600))
		log.Printf("区块时间timeStamp: %s ", timeStamp)
		err = h.repository.RecordBalanceChange(h.config.ChainID, transferEvent.From, vLog.TxHash,
			vLog.BlockNumber, transferEvent.Value.Uint64(), balanceFrom.String(), evenType, timeStamp)
		if err != nil {
			return err
		}

		err = h.repository.UpdateUserBalance(h.config.ChainID, transferEvent.To, balanceTo.Uint64())
		if err != nil {
			return err
		}
		err = h.repository.RecordBalanceChange(h.config.ChainID, transferEvent.To, vLog.TxHash,
			vLog.BlockNumber, transferEvent.Value.Uint64(), balanceTo.String(), evenType, timeStamp)
		if err != nil {
			return err
		}

		err = h.repository.SaveChain(h.config.Name, h.config.ChainID, h.config.ContractAddr, end)
		if err != nil {
			return err
		}

	}
	return nil
}

// 获取用户在指定ERC-20代币合约中的余额
func getUserTokenBalance(ctx context.Context, client *ethclient.Client, tokenContract common.Address, userAddress common.Address) (*big.Int, uint8, string, error) {
	// 解析ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return nil, 0, "", fmt.Errorf("解析ABI失败: %w", err)
	}

	// 调用balanceOf方法
	data, err := parsedABI.Pack("balanceOf", userAddress)
	if err != nil {
		return nil, 0, "", fmt.Errorf("打包调用数据失败: %w", err)
	}

	// 执行合约调用
	msg := ethereum.CallMsg{
		To:   &tokenContract,
		Data: data,
	}
	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, 0, "", fmt.Errorf("调用合约失败: %w", err)
	}

	// 解析余额
	var balance *big.Int
	if err := parsedABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, 0, "", fmt.Errorf("解析余额失败: %w", err)
	}

	// 获取代币小数位
	decimalsData, err := parsedABI.Pack("decimals")
	if err != nil {
		return balance, 18, "", fmt.Errorf("打包decimals调用数据失败: %w", err)
	}
	decimalsResult, err := client.CallContract(ctx, ethereum.CallMsg{To: &tokenContract, Data: decimalsData}, nil)
	var decimals uint8
	if err != nil {
		return balance, 18, "", fmt.Errorf("获取小数位失败，使用默认18位: %w", err)
	}
	if err := parsedABI.UnpackIntoInterface(&decimals, "decimals", decimalsResult); err != nil {
		return balance, 18, "", fmt.Errorf("解析小数位失败，使用默认18位: %w", err)
	}

	// 获取代币符号
	symbolData, err := parsedABI.Pack("symbol")
	if err != nil {
		return balance, decimals, "", fmt.Errorf("打包symbol调用数据失败: %w", err)
	}
	symbolResult, err := client.CallContract(ctx, ethereum.CallMsg{To: &tokenContract, Data: symbolData}, nil)
	var symbol string
	if err != nil {
		return balance, decimals, "未知代币", fmt.Errorf("获取代币符号失败: %w", err)
	}
	if err := parsedABI.UnpackIntoInterface(&symbol, "symbol", symbolResult); err != nil {
		return balance, decimals, "未知代币", fmt.Errorf("解析代币符号失败: %w", err)
	}

	return balance, decimals, symbol, nil
}
