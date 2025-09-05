package db

import (
	. "POINTSTOKEN/types"
	"database/sql"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"time"
)

type Repository interface {
	//链相关操作
	SaveChain(chainName string, chainID uint64, contractAddr string, lastBlock uint64) error
	GetChains() ([]uint64, error)
	UpdateChainLastBlock(chainID uint64, lastBlock uint64) error
	GetChainLastBlock(chainID uint64) (uint64, error)

	// 余额相关操作
	UpdateUserBalance(chainID uint64, userAddr common.Address, balance uint64) error
	GetUserBalance(chainID uint64, userAddr common.Address) (uint64, error)
	RecordBalanceChange(chainID uint64, userAddr common.Address, txHash common.Hash,
		blockNumber uint64, changeAmount uint64, balanceAfter string, eventType string, timeStamp time.Time) error
	GetBalanceChange(chainID uint64) ([]UserBalanceChange, error)
	GetAllUserBalance() ([]UserBalance, error)

	// 积分相关操作
	UpdateUserPoints(chainId uint64, userAddr common.Address, addPoints string) (string, error)
	GetUserPoints(chainId uint64, userAddr common.Address) (string, error)
	RecordPointsCalculation(chainId uint64, userAddr common.Address, calculation time.Time,
		balance string, pointsAdd string, totalAfter string) error
	GetPointLastRecordTime(chainId uint64, userAddr common.Address) (time.Time, error)
}

type UserBalance struct {
	ChainID  uint64
	UserAddr common.Address
	Balance  BigInt
}
type UserBalanceChange struct {
	ChainID       uint64
	UserAddr      common.Address
	BalanceChange BigInt
	BalanceAfter  BigInt
	EventType     string
	CreatedAt     time.Time
}

type DBRepository struct {
	Db *DB
}

// NewDBRepository 创建新的数据库操作实例
func NewDBRepository(db *DB) *DBRepository {
	return &DBRepository{Db: db}
}

// SaveChain 保存链信息
func (r *DBRepository) SaveChain(chainName string, chainID uint64, contractAddr string, lastBlock uint64) error {
	_, err := r.Db.Exec(`
		INSERT INTO chains (name, chain_id, contract_addr, last_processed_block) 
		VALUES (?, ?, ?,?) ON DUPLICATE KEY UPDATE last_processed_block = ?`, chainName, chainID, contractAddr, lastBlock, lastBlock)
	return err
}

// GetChains 获取保存的链
func (r *DBRepository) GetChains() ([]uint64, error) {
	var chains []uint64
	rows, err := r.Db.Query(`
		select chain_id from chains `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chainId uint64
	for rows.Next() {
		err = rows.Scan(&chainId)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chainId)
	}
	return chains, rows.Err()
}

// UpdateChainLastBlock 更新链的最后处理区快
func (r *DBRepository) UpdateChainLastBlock(chainID uint64, lastBlock uint64) error {
	_, err := r.Db.Exec(`
		UPDATE chains SET last_processed_block = ?, updated_at = NOW() 
		where chain_id = ? `, lastBlock, chainID)
	return err
}

// GetChainLastBlock 获取链的最后处理块
func (r *DBRepository) GetChainLastBlock(chainID uint64) (uint64, error) {
	var lastBlock uint64
	err := r.Db.QueryRow(`
		select last_processed_block from chains where chain_id = ?`, chainID).Scan(&lastBlock)
	return lastBlock, err
}

// UpdateUserBalance 更新用户余额
func (r *DBRepository) UpdateUserBalance(chainID uint64, userAddr common.Address, balance uint64) error {
	_, err := r.Db.Exec(`
		insert into user_balances (chain_id, user_addr, balance) 
		values (?, ?, ?) ON DUPLICATE KEY UPDATE balance = ?`, chainID, userAddr.Hex(), balance, balance)
	return err
}

// GetUserBalance 获取用户余额
func (r *DBRepository) GetUserBalance(chainID uint64, userAddr common.Address) (uint64, error) {
	var balance uint64
	err := r.Db.QueryRow(`
		select balance from user_balances where chain_id = ? and user_addr = ?`,
		chainID, userAddr.Hex()).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return balance, err
}

// RecordBalanceChange 记录余额变动
func (r *DBRepository) RecordBalanceChange(chainID uint64, userAddr common.Address, txHash common.Hash,
	blockNumber uint64, changeAmount uint64, balanceAfter string, eventType string, timeStamp time.Time) error {
	_, err := r.Db.Exec(`
		insert into balance_changes (
		chain_id, user_addr, transaction_hash, block_number, change_amount, balance_after, event_type, created_at) 
		values (?,?,?,?,?,?,?,?)`, chainID, userAddr.Hex(), txHash.Hex(), blockNumber, changeAmount, balanceAfter, eventType, timeStamp)
	return err
}

// GetBalanceChange 获取余额变动信息
func (r *DBRepository) GetBalanceChange(chainID uint64) ([]UserBalanceChange, error) {
	rows, err := r.Db.Query(`
		select chain_id, user_addr, change_amount, balance_after,event_type,created_at 
		from balance_changes where chain_id = ? order by chain_id,user_addr`, chainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var userBalanceChange []UserBalanceChange
	for rows.Next() {
		var ubc UserBalanceChange
		var addrStr string
		err = rows.Scan(&ubc.ChainID, &addrStr, &ubc.BalanceChange, &ubc.BalanceAfter, &ubc.EventType, &ubc.CreatedAt)
		if err != nil {
			return nil, err
		}
		ubc.UserAddr = common.HexToAddress(addrStr)
		userBalanceChange = append(userBalanceChange, ubc)
	}
	return userBalanceChange, rows.Err()
}

// GetAllUserBalance 获取所以用户余额
func (r *DBRepository) GetAllUserBalance() ([]UserBalance, error) {
	rows, err := r.Db.Query(`select chain_id, user_addr, balance from user_balances`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userBalances []UserBalance
	for rows.Next() {
		var ub UserBalance
		var addrStr string
		err = rows.Scan(&ub.ChainID, &addrStr, &ub.Balance)
		if err != nil {
			return nil, err
		}
		ub.UserAddr = common.HexToAddress(addrStr)
		userBalances = append(userBalances, ub)
	}
	return userBalances, rows.Err()
}

func (r *DBRepository) UpdateUserPoints(chainId uint64, userAddr common.Address, addPoints string) (string, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	//更新积分
	_, err = tx.Exec(`
		insert into user_points (chain_id, user_addr, total_points) values (?, ?, ?) 
		ON DUPLICATE KEY UPDATE total_points = ?
		`, chainId, userAddr.Hex(), addPoints, addPoints)
	if err != nil {
		return "", err
	}

	var finalPoints string
	err = tx.QueryRow(`
		select total_points from user_points where user_addr = ?  and chain_id = ?`, userAddr.Hex(), chainId).Scan(&finalPoints)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return finalPoints, nil
}

func (r *DBRepository) GetUserPoints(chainId uint64, userAddr common.Address) (string, error) {
	var points string
	err := r.Db.QueryRow(`
		select total_points from user_points where user_addr = ?  and chain_id = ?`, userAddr.Hex(), chainId).Scan(&points)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "0", err
	}
	if points == "" {
		points = "0"
	}
	return points, nil
}

// RecordPointsCalculation 记录积分计算
func (r *DBRepository) RecordPointsCalculation(chainId uint64, userAddr common.Address, calculatedAt time.Time,
	balance, pointsAdded, totalAfter string) error {
	_, err := r.Db.Exec(`
		insert into points_calculations (chain_id, user_addr, calculated_at, balance, points_added, total_points_after) 
		values (?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE points_added = ?`, chainId, userAddr.Hex(), calculatedAt, balance, pointsAdded, totalAfter, pointsAdded)
	return err
}

func (r *DBRepository) GetPointLastRecordTime(chainId uint64, userAddr common.Address) (time.Time, error) {
	var lastTime time.Time
	err := r.Db.QueryRow(`
		select calculated_at from points_calculations 
		where user_addr = ? and chain_id = ? order by calculated_at desc`, userAddr.Hex(), chainId).Scan(&lastTime)
	if err != nil {
		return lastTime, err
	}
	return lastTime, nil
}
