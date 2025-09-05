package main

import (
	"POINTSTOKEN/config"
	"POINTSTOKEN/db"
	"POINTSTOKEN/service"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 加载配置
	cfgPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfigFile(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	database, err := db.NewDB(cfg.Database)
	if err != nil {
		return
	}
	defer database.Close()

	dbRepo := db.NewDBRepository(database)

	manager, err := service.NewChainManager(cfg.Chains, dbRepo)
	if err != nil {
		return
	}

	// 启动事件监听
	ctx, cancel := context.WithCancel(context.Background())
	manager.StartEventListeners(ctx)

	// 初始化积分计算器
	pointCalculator := service.NewPointsCalculator(&cfg.Points, dbRepo)
	// 启动定时积分计算任务
	err = pointCalculator.Start()
	if err != nil {
		return
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭
	log.Println("Shutting down service...")
	cancel()
	pointCalculator.Stop()
	time.Sleep(5 * time.Second)
	log.Println("Service stopped")
}
