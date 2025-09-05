package config

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Chains   []ChainConfig  `mapstructure:"chains"`
	Points   PointsConfig   `mapstructure:"points"`
}

type DatabaseConfig struct {
	DSN          string        `mapstructure:"dsn"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
}

type ChainConfig struct {
	Name          string        `mapstructure:"name"`
	ChainID       uint64        `mapstructure:"chain_id"`
	RPCUrl        string        `mapstructure:"rpc_url"`
	ContractAddr  string        `mapstructure:"contract_addr"`
	StartBlock    uint64        `mapstructure:"start_block"`
	Confirmations uint64        `mapstructure:"confirmations"`
	PollInterval  time.Duration `mapstructure:"poll_interval"`
}

type PointsConfig struct {
	Rate     float64 `mapstructure:"rate"`      //积分计算比例
	CronSpec string  `mapstructure:"cron_spec"` //定时任务表达式
}

// LoadConfigFile 加载配置文件
func LoadConfigFile(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	//设置默认值
	for i := range config.Chains {
		if config.Chains[i].Confirmations == 0 {
			config.Chains[i].Confirmations = 6
		}
		if config.Chains[i].PollInterval == 0 {
			config.Chains[i].PollInterval = 30 * time.Second
		}
	}
	if config.Points.Rate == 0 {
		config.Points.Rate = 0.05
	}
	if config.Points.CronSpec == "" {
		config.Points.CronSpec = "0 * * * *" //每小时执行一次
	}
	return &config, nil
}

// GetContractAddress 返回合约地址
func (c *ChainConfig) GetContractAddress() common.Address {
	return common.HexToAddress(c.ContractAddr)
}
