package utilities

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	// pgx replaces lib/pq because lib/pq uses unnamed prepared statements
	// that PgBouncer cannot track in transaction-pool mode — pgx uses named
	// prepared statements which PgBouncer ≥1.21 handles correctly.
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// pgDriverName is the driver name registered by pgx/v5/stdlib.
// We keep DBConfiguration.DbType = "postgres" for backward compatibility
// (callers, env vars, and connection-string format unchanged) and translate
// to "pgx" only when calling sql.Open.
const pgDriverName = "pgx"

type DBParam struct {
	Host     string
	Port     string
	Name     string
	Schema   string
	User     string
	Password string
	AppName  string
	Timeout  int
	MaxOpen  int
	MaxIdle  int
	Logging  bool
}

type DBConfiguration struct {
	DbType         string
	Host           string
	Port           string
	Schema         string
	DBName         string
	Username       string
	Password       string
	Logging        bool
	SessionName    string
	ConnectTimeOut int
	MaxOpenConn    int
	MaxIdleConn    int
	Migrate        bool
	PreparedStmt   bool
}

const (
	Mysql      = "mysql"
	Postgresql = "postgres"
)

func GetDBConfig() DBConfiguration {
	cfg := DBConfiguration{
		Host:           EnvString("DB_HOST"),
		DBName:         EnvString("DB_NAME"),
		Username:       EnvString("DB_USERNAME"),
		Password:       EnvString("DB_PASSWORD"),
		Logging:        EnvBool("DB_DEBUG"),
		Port:           EnvString("DB_PORT"),
		Schema:         EnvString("DB_SCHEMA"),
		SessionName:    EnvString("DB_SESSION_NAME"),
		ConnectTimeOut: EnvInt("DB_CONNECT_TIMEOUT"),
		MaxOpenConn:    EnvInt("DB_MAX_OPEN_CONN"),
		MaxIdleConn:    EnvInt("DB_MAX_IDLE_CONN"),
	}

	//default db connection time out
	if cfg.ConnectTimeOut == 0 {
		cfg.ConnectTimeOut = 30
	}
	//default db maximum open connection
	if cfg.MaxOpenConn == 0 {
		cfg.MaxOpenConn = 50
	}
	//default db maximum idle connection
	if cfg.MaxIdleConn == 0 {
		cfg.MaxIdleConn = 10
	}

	return cfg
}

func ConnectDB(cfg DBConfiguration) (*sql.DB, error) {
	connString := makeConnString(cfg.DbType, cfg.SessionName, cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.ConnectTimeOut)
	driver := cfg.DbType
	if cfg.DbType == Postgresql {
		driver = pgDriverName
	}
	sql, err := sql.Open(driver, connString)
	if err != nil {
		return nil, err
	}

	sql.SetMaxIdleConns(cfg.MaxIdleConn)
	sql.SetMaxOpenConns(cfg.MaxOpenConn)
	sql.SetConnMaxLifetime(time.Hour)

	return sql, nil
}

func InitGorm(sqlConn *sql.DB, cfg DBConfiguration) (*gorm.DB, error) {
	dbLogger := CreateLogger(cfg.Logging)
	// Use the dedicated PreparedStmt flag instead of (mistakenly) reusing Logging.
	// Previously the GORM PrepareStmt cache was toggled by DB_DEBUG, which conflated
	// SQL logging with statement caching.
	db, err := NewGormDB(cfg.DbType, sqlConn, dbLogger, cfg.PreparedStmt)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func makeConnString(dbtype, conName, host, port, user, dbname, password string, timeOut int) string {
	if dbtype == Postgresql {
		return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s connect_timeout=%d application_name=%s",
			host, port, user, dbname, password, timeOut, conName)
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)
}

func CreateLogger(debug bool) logger.Interface {
	level := logger.Silent
	if debug {
		level = logger.Info
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  level,       // Log level
			IgnoreRecordNotFoundError: false,       // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
}

func NewGormDB(dbtype string, conn *sql.DB, log logger.Interface, preparedStatement bool) (*gorm.DB, error) {
	if dbtype == Mysql {
		return gorm.Open(mysql.New(mysql.Config{
			Conn:                      conn,
			SkipInitializeWithVersion: true,
		}), &gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 log,
			PrepareStmt:            preparedStatement,
		})
	}

	return gorm.Open(postgres.New(
		postgres.Config{Conn: conn}),
		&gorm.Config{Logger: log, PrepareStmt: preparedStatement})
}
