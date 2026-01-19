// Package ioc provides database initialization.
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

// InitDB initializes GORM database connection from config.
func InitDB(cfg *config.Config, zapLogger *zap.Logger) (*gorm.DB, error) {
	if !cfg.MySQL.Enabled {
		zapLogger.Info("MySQL is disabled, skipping database initialization")
		return nil, nil
	}

	// Build DSN from config
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database,
		cfg.MySQL.Charset,
	)

	// Configure GORM logger
	logLevel := logger.Silent
	if cfg.Log.Level == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate tables
	if err := db.AutoMigrate(
		&dao.Provider{},
		&dao.RoutingRule{},
		&dao.APIKey{},
		&dao.LoadBalanceGroup{},
		&dao.LoadBalanceMember{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	zapLogger.Info("database initialized",
		zap.String("host", cfg.MySQL.Host),
		zap.Int("port", cfg.MySQL.Port),
		zap.String("database", cfg.MySQL.Database),
		zap.String("dsn", maskDSN(dsn)),
	)

	return db, nil
}

// maskDSN masks the password in DSN for logging.
func maskDSN(dsn string) string {
	// Find password part between first : and @
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
