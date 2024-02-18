package database

import (
	"github.com/gerdooshell/tax-logger/lib/cache/lrucache"
	"github.com/gerdooshell/tax-logger/lib/database/config"
	dbEngine "github.com/gerdooshell/tax-logger/lib/database/engine"
	dbHosting "github.com/gerdooshell/tax-logger/lib/database/hosting"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

type ConnectionPool struct {
	dbHosting.HostTag
	dbEngine.Engine
	Conn *gorm.DB
}

var connectionsCache = lrucache.NewLRUCache[struct {
	Engine  dbEngine.Engine
	HostTag dbHosting.HostTag
}](5)

var mu sync.Mutex

func NewConnectionPool(config config.ConnectionConfig) (*ConnectionPool, error) {
	mu.Lock()
	defer mu.Unlock()
	var connPool *ConnectionPool
	poolGeneric, err := connectionsCache.Read(config.GetSignature())
	if err == nil {
		connPool = poolGeneric.(*ConnectionPool)
		return connPool, nil
	}
	conn, err := gorm.Open(postgres.Open(config.GetConnectionString()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	db, _ := conn.DB()
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Hour * 24)
	connPool = &ConnectionPool{
		HostTag: config.GetHostTag(),
		Engine:  config.GetEngine(),
		Conn:    conn,
	}
	connectionsCache.Add(config.GetSignature(), connPool)
	return connPool, nil
}
