// Package ioc 提供数据库初始化。
package ioc

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ai-gateway/config"
	"ai-gateway/internal/repository/dao"
)

// InitDB 根据配置初始化 GORM 数据库连接。
func InitDB(cfg *config.Config, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 从配置构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
		cfg.MySQL.Charset,
	)

	// 配置 GORM 日志记录器
	logLevel := logger.Silent
	if cfg.Log.Level == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("无法连接数据库: %w", err)
	}

	// 获取底层 SQL DB 以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("无法获取 sql.DB: %w", err)
	}

	// 设置连接池设置
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移表
	if err := db.AutoMigrate(
		&dao.Provider{},
		&dao.RoutingRule{},
		&dao.APIKey{},
		&dao.LoadBalanceGroup{},
		&dao.LoadBalanceMember{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	zapLogger.Info("database initialized",
		zap.String("host", cfg.MySQL.Host),
		zap.Int("port", cfg.MySQL.Port),
		zap.String("database", cfg.MySQL.Database),
		zap.String("dsn", maskDSN(dsn)),
	)

	return db, nil
}

// maskDSN 屏蔽 DSN 中的密码以进行日志记录。
func maskDSN(dsn string) string {
	// 在第一个 : 和 @ 之间找到密码部分
	parts := strings.SplitN(dsn, ":", 2)
	if len(parts) < 2 {
		return dsn
	}

	afterColon := parts[1]
	atIndex := strings.Index(afterColon, "@")
	if atIndex == -1 {
		return dsn
	}

	return parts[0] + ":****" + afterColon[atIndex:]
}
