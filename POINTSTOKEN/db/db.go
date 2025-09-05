package db

import (
	"POINTSTOKEN/config"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	*sql.DB
}

func NewDB(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// 解析连接超时时间
	//timeout, err := time.ParseDuration(strconv.FormatInt(int64(cfg.ConnTimeout), 10))
	//if err != nil {
	//	return nil, err
	//}
	//db.SetConnMaxLifetime(timeout)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
