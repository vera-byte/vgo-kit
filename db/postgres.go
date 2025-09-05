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
	"github.com/vera-byte/vgo-kit/logger"
	"go.uber.org/zap"
)

// PostgresStore PostgreSQL存储
// 提供统一的 *sql.DB 与 dbr.Session 访问入口，确保仅使用单一连接池
// 字段:
//   - DB: 标准库数据库连接（连接池）
//   - Session: dbr会话对象（基于同一*sql.DB，不创建新的连接池）
type PostgresStore struct {
	DB      *sql.DB
	Session *dbr.Session
}

// NewPostgresStore 创建PostgreSQL存储实例
// 参数:
//   - dsn: 数据源名称（支持postgres://或关键字形式）
// 返回:
//   - *PostgresStore: 存储实例
//   - error: 错误信息
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	// 标准化DSN前缀（仅将 postgresql:// 规范为 postgres://，其余格式保持不变）
	if strings.HasPrefix(dsn, "postgresql://") {
		dsn = "postgres://" + strings.TrimPrefix(dsn, "postgresql://")
	}

	// 通过 dbr.Open 创建连接（内部持有单一 *sql.DB 连接池）
	conn, err := dbr.Open("postgres", dsn, nil)
	if err != nil {
		return nil, fmt.Errorf("db open failed: %w", err)
	}

	// 带超时的连接测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if pingErr := conn.DB.PingContext(ctx); pingErr != nil {
		_ = conn.DB.Close()
		return nil, fmt.Errorf("db ping failed: %w", pingErr)
	}

	// 优化连接池配置
	conn.DB.SetMaxOpenConns(25)                  // 建议值: (2 * CPU核心数) + 备用连接
	conn.DB.SetMaxIdleConns(5)                   // 小于MaxOpenConns
	conn.DB.SetConnMaxLifetime(30 * time.Minute) // 避免云服务断开
	conn.DB.SetConnMaxIdleTime(5 * time.Minute)  // 主动回收闲置连接

	// 启动连接健康监控（仅记录告警，不影响业务流程）
	go monitorConnection(conn.DB, 10*time.Second, vgokit.Log)

	return &PostgresStore{
		DB:      conn.DB,
		Session: conn.NewSession(nil),
	}, nil
}

// 连接健康监控
// monitorConnection 监控数据库连接健康状态
// 参数:
//   - db: 数据库连接实例
//   - interval: 监控间隔时间
//   - logger: 日志记录器接口
func monitorConnection(db *sql.DB, interval time.Duration, logger logger.Logger) {
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
// 返回: error 关闭过程中产生的错误
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

	// 仅关闭底层 *sql.DB 连接池；dbr.Session 基于该连接池，无需单独关闭
	closeFunc("DB", s.DB.Close)

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
