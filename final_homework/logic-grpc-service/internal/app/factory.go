package app

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const DefaultMySQLDSN = "root:123456@tcp(127.0.0.1:3306)/smart_recruitment?parseTime=true&charset=utf8mb4&loc=Local"

func NewServiceFromEnv(mysqlDSN string) (*Service, error) {
	cfg, err := ConfigFromFileAndEnv("")
	if err != nil {
		return nil, err
	}
	if mysqlDSN != "" {
		cfg.MySQL.DSN = mysqlDSN
	}
	cosCfg, cosErr := NormalizeCOSConfig(cfg.COS)
	var cosStore *COSStore
	if cosErr == nil {
		cosStore, cosErr = NewCOSStore(cosCfg)
	}
	aiProvider, aiErr := NewEinoAIProvider(context.Background(), cfg.AI)
	if cfg.MySQL.DSN == "" {
		cfg.MySQL.DSN = DefaultMySQLDSN
	}
	db, err := sql.Open("mysql", cfg.MySQL.DSN)
	if err != nil {
		return nil, fmt.Errorf("初始化 MySQL 连接失败: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	service, err := NewServiceWithMySQL(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if cosErr == nil {
		service.WithCOS(cosStore)
	} else {
		service.cosErr = cosErr
		service.mysql.cosErr = cosErr
	}
	if aiErr == nil {
		service.WithAI(aiProvider)
	} else {
		service.aiErr = aiErr
		service.mysql.aiErr = aiErr
	}
	return service, nil
}
