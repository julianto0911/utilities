package utilities

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

func ConnectDB(cfg DBConfiguration) (*sql.DB, error) {
	connString := makeConnString(cfg.DbType, cfg.SessionName, cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.ConnectTimeOut)
	sql, err := sql.Open(cfg.DbType, connString)
	if err != nil {
		return nil, err
	}

	sql.SetMaxIdleConns(cfg.MaxIdleConn)
	sql.SetMaxOpenConns(cfg.MaxOpenConn)
	sql.SetConnMaxLifetime(time.Hour)

	return sql, nil
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
