package service

import (
	"POINTSTOKEN/config"
	"POINTSTOKEN/db"
	. "POINTSTOKEN/types"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/robfig/cron/v3"
	"log"
	"math/big"
	"time"
)

var ZeroAddress common.Address = common.HexToAddress("0x0000000000000000000000000000000000000000")

type PointsCalculator struct {
	db       *db.DBRepository
	cron     *cron.Cron
	pointCfg *config.PointsConfig
	entryID  cron.EntryID
	running  bool
}

func NewPointsCalculator(pointCfg *config.PointsConfig, db *db.DBRepository) *PointsCalculator {
	// 配置 cron 解析器，支持可选的秒字段（6字段或5字段格式）
	cronParser := cron.NewParser(
		cron.SecondOptional |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow |
			cron.Descriptor,
	)

	return &PointsCalculator{
		pointCfg: pointCfg,
		db:       db,
		cron:     cron.New(cron.WithParser(cronParser)),
	}
}

func (p *PointsCalculator) Start() error {
	if p.running {
		return fmt.Errorf("积分计算服务已在运行")
	}
	p.running = true

	//添加定时任务
	entryID, err := p.cron.AddFunc(p.pointCfg.CronSpec, p.calculatePoints)
	if err != nil {
		return fmt.Errorf("添加定时任务失败: %v", err)
	}
	p.entryID = entryID

	// 立即执行一次，然后按计划运行
	go p.calculatePoints()

	p.cron.Start()
	log.Println("积分计算服务已启动")
	return nil
}

func (p *PointsCalculator) Stop() {
	if !p.running {
		return
	}
	p.cron.Remove(p.entryID)
	p.cron.Stop()
	p.running = false
	log.Println("积分计算服务已停止")
}

func (p *PointsCalculator) calculatePoints() {
	log.Println("开始执行积分计算...")
	chains, err := p.db.GetChains()
	if err != nil {
		log.Printf("获取链信息失败: %v", err)
	}

	for _, chain := range chains {
		if err := p.calculatePointsForChain(chain); err != nil {
			log.Printf("为链 %s 计算积分失败: %v", chain, err)
		}
	}
}

func (p *PointsCalculator) calculatePointsForChain(chain uint64) error {
	log.Printf("")
	log.Printf("=======================")
	log.Printf("==========%v===========", chain)
	log.Printf("=======================")
	users, err := p.db.GetBalanceChange(chain)
	if err != nil {
		return fmt.Errorf("获取用户余额失败: %v", err)
	}
	if len(users) == 0 {
		log.Printf("链 %s 上没有需要计算积分的用户", chain)
		return nil
	}

	tx, err := p.db.Db.Begin()
	if err != nil {
		return fmt.Errorf("创建事务失败: %v", err)
	}
	defer tx.Rollback()
	log.Printf("链 %v 上用户: %s ", chain, users)

	userAddr := users[0].UserAddr
	addPoints := big.NewInt(0)
	var startTime time.Time
	var endTime time.Time
	var balanceAfter BigInt
	for i, user := range users {
		if userAddr == ZeroAddress {
			userAddr = user.UserAddr
			continue
		}
		log.Printf("用户地址：%s", users[i].UserAddr.String())
		isLast := i == len(users)-1
		//获取上一次计算积分的时间
		lastTime, err := p.db.GetPointLastRecordTime(chain, users[i].UserAddr)

		//有交易记录，还未开始计算积分
		if lastTime.IsZero() && !isLast {
			startTime = user.CreatedAt
			endTime = users[i+1].CreatedAt
			balanceAfter = user.BalanceAfter
		}
		//计算最后一条转账后到当前时间的积分
		if isLast {
			startTime = user.CreatedAt
			endTime = time.Now()
			balanceAfter = user.BalanceAfter
		}
		//若该条交易记录的创建时间大于上一次计算积分的时间，则从上一次计算积分的时间开始计算积分
		log.Printf("CreatedAt:%s  ,lastTime:%s", user.CreatedAt, lastTime)
		if user.CreatedAt.After(lastTime) && !lastTime.IsZero() {
			startTime = lastTime
			endTime = user.CreatedAt
			balanceAfter = user.BalanceAfter
		}

		addPoints, err = p.calculate(addPoints, startTime, endTime, balanceAfter)
		if err != nil {
			return fmt.Errorf("计算用户积分失败: %v", err)
		}

		log.Printf("用户地址积分：%s", addPoints.String())
		flag := !isLast && user.UserAddr != users[i+1].UserAddr
		if isLast || flag {

			points, err := p.db.GetUserPoints(chain, userAddr)
			if err != nil {
				return fmt.Errorf("获取用户积分失败: %v", err)
			}
			pointsF := new(big.Int)
			pointsF.SetString(points, 10)

			totalPointsAfter := big.NewInt(0)
			totalPointsAfter.Add(addPoints, pointsF)

			log.Printf("链:%v， 地址：%s, 用户积分:%s", chain, userAddr.String(), totalPointsAfter.String())
			value, err := p.db.UpdateUserPoints(chain, userAddr, totalPointsAfter.String())
			if err != nil {
				return err
			}
			log.Println("用户总积分:s%", value)
			if addPoints == big.NewInt(0) {
				continue
			}
			err = p.db.RecordPointsCalculation(chain, userAddr, user.CreatedAt,
				user.BalanceAfter.ToBigInt().String(),
				addPoints.String(),
				totalPointsAfter.String())
			if err != nil {
				return err
			}

			lastTime = user.CreatedAt
			userAddr = user.UserAddr
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (p *PointsCalculator) calculate(addPoints *big.Int, startTime time.Time, endTime time.Time, balanceAfter BigInt) (*big.Int, error) {
	diff := endTime.Sub(startTime)
	// 转换为小时（多种精度）
	hours := diff.Hours()

	// 1. 处理小数乘法：p.pointCfg.Rate（小数）* hours * 1e18（需转为分数避免精度损失）
	// 示例：若 Rate=0.01（即 1/100），先转为分子=1、分母=100
	rateNumerator := big.NewInt(int64(p.pointCfg.Rate * 1000)) // 假设 Rate 保留2位小数，放大1000倍为整数
	rateDenominator := big.NewInt(1000)                        // 对应分母
	log.Printf("小时：%f，rate ：%s, balanceAfter: %s", hours, rateNumerator.String(), balanceAfter.ToBigInt().String())
	// 2. 处理 hours（假设是 int64）
	sec := float64(60 * 60 * 1e9)
	hoursBig0 := hours * sec
	log.Printf("rateNumerator:%s  ,hoursBig0:%f", rateNumerator.String(), hoursBig0)
	hoursBig := big.NewInt(int64(hoursBig0))
	bSec := big.NewInt(60 * 60 * 1e9)
	log.Printf("rateNumerator:%s  ,hoursBig:%s", rateNumerator.String(), hoursBig.String())
	tempMul1 := new(big.Int).Mul(rateNumerator, hoursBig)
	tempMul2 := new(big.Int).Mul(bSec, rateDenominator)

	balanceAfterPtr := balanceAfter // 关键：值类型转指针类型
	log.Printf("balanceAfter:%s  ,tempMul1:%s", balanceAfter.ToBigInt().String(), tempMul1.String())
	mulResult := new(big.Int).Mul(balanceAfterPtr.ToBigInt(), tempMul1)
	bigInt := new(big.Int).Div(mulResult, tempMul2)
	log.Printf("mulResult:%s  ,tempMul2:%s", mulResult.String(), tempMul2.String())
	// 6. 累加至 addPoints（假设 addPoints 是 *big.Int 类型，需初始化）
	if addPoints == nil {
		addPoints = new(big.Int) // 初始化空的 big.Int 指针
	}

	addPoints.Add(addPoints, bigInt) // 累加：addPoints = addPoints + mulResult
	log.Printf("addPoints:%s  ,bigInt:%s", addPoints.String(), bigInt.String())
	return addPoints, nil
}
