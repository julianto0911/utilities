package utilities

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBManager struct {
	connections map[string]*gorm.DB
	mu          sync.RWMutex
	config      DBConfiguration
}

// Multi database connection manager
func NewDBManager(defaultConfig DBConfiguration) *DBManager {
	return &DBManager{
		connections: make(map[string]*gorm.DB),
		config:      defaultConfig,
	}
}

/*
Use context to pass value of database and db username
example :
ctx.WithValue("dbname", "db1")
ctx.WithValue("dbuser", "user1")
*/

func (m *DBManager) DB(ctx context.Context) (*gorm.DB, error) {
	dbName := fmt.Sprintf("%s", ctx.Value("dbname"))
	dbUser := fmt.Sprintf("%s", ctx.Value("dbuser"))
	if dbName == "" {
		return nil, fmt.Errorf("database name is required")
	}

	if dbUser == "" {
		return nil, fmt.Errorf("database username is required")
	}

	m.mu.RLock()
	db, exists := m.connections[dbName]
	m.mu.RUnlock()

	if exists {
		return db, nil
	}

	return m.createConnection(dbName, dbUser)
}

func (m *DBManager) createConnection(dbname, dbuser string) (*gorm.DB, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check if connection was created while waiting for lock
	if db, exists := m.connections[dbname]; exists {
		return db, nil
	}

	// Clone base config and modify for tenant
	dbConfig := m.config
	dbConfig.DBName = dbname
	dbConfig.Username = dbuser

	// Create new connection
	sqlConn, err := ConnectDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}

	// Initialize GORM
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlConn,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	m.connections[dbname] = db
	return db, nil
}

// Close all connections before exit application
func (m *DBManager) CloseConnections() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for tenantID, db := range m.connections {
		sqlDB, err := db.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("tenant %s: %w", tenantID, err))
			continue
		}
		if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("tenant %s: %w", tenantID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}
