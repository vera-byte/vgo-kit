package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gocraft/dbr/v2"
	_ "github.com/lib/pq"
	vgokit "github.com/vera-byte/vgo-kit"
	"go.uber.org/zap"
)

// PostgresStore PostgreSQL存储
type PostgresStore struct {
	DB      *sql.DB
	Session *dbr.Session
}

// NewPostgresStore 创建PostgreSQL存储实例
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	// 标准化DSN格式
	if !strings.Contains(dsn, "://") {
		dsn = "postgres://" + strings.TrimPrefix(dsn, "postgresql://")
	}

	// 创建标准连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open failed: %w", err)
	}

	// 带超时的连接测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if pingErr := db.PingContext(ctx); pingErr != nil {
		db.Close()
		return nil, fmt.Errorf("db ping failed: %w", pingErr)
	}

	// 优化连接池配置
	db.SetMaxOpenConns(25)                  // 建议值: (2 * CPU核心数) + 备用连接
	db.SetMaxIdleConns(5)                   // 小于MaxOpenConns
	db.SetConnMaxLifetime(30 * time.Minute) // 避免云服务断开
	db.SetConnMaxIdleTime(5 * time.Minute)  // 主动回收闲置连接

	// 创建dbr连接
	conn, err := dbr.Open("postgres", dsn, nil)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("dbr open failed: %w", err)
	}

	// 启动连接监控

	go monitorConnection(db, 10*time.Second, vgokit.Log.Logger)

	return &PostgresStore{
		DB:      db,
		Session: conn.NewSession(nil),
	}, nil
}

// 连接健康监控
func monitorConnection(db *sql.DB, interval time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := db.PingContext(ctx)
		cancel()

		if err != nil && logger != nil {
			logger.Warn("Database connection unhealthy", zap.Error(err))
		}
	}
}

// Close 关闭数据库连接（带重试机制）
func (s *PostgresStore) Close() error {
	const maxRetries = 3
	var errs []error

	closeFunc := func(name string, closer func() error) {
		for i := 0; i < maxRetries; i++ {
			if err := closer(); err == nil {
				return
			}
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
		errs = append(errs, fmt.Errorf("failed to close %s after %d retries", name, maxRetries))
	}

	closeFunc("DB", s.DB.Close)
	closeFunc("dbr session", s.Session.Close)

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
