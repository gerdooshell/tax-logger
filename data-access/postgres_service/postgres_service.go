package postgresService

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gerdooshell/tax-logger/data-access/postgres_service/relations"
	"github.com/gerdooshell/tax-logger/entities"
	"github.com/gerdooshell/tax-logger/environment"
	"github.com/gerdooshell/tax-logger/interactors"
	"github.com/gerdooshell/tax-logger/lib/database"
	"github.com/gerdooshell/tax-logger/lib/database/config"

	"gorm.io/gorm/clause"
)

type postgresServiceImpl struct {
	engine *database.ConnectionPool
	conf   config.ConnectionConfig
	mu     sync.Mutex
}

var postgresServiceSingleton *postgresServiceImpl

func NewPostgresService() interactors.DataAccess {
	if postgresServiceSingleton == nil {
		postgresServiceSingleton = &postgresServiceImpl{
			engine: nil,
			conf:   nil,
		}
	}
	go postgresServiceSingleton.retryConnection()
	return postgresServiceSingleton
}

func (pg *postgresServiceImpl) retryConnection() {
	// in case of losing an established connection, orm will reconnect automatically
	// in case of first connection failure, this function retries
	const retryWaiting = time.Second * 20
	tick := time.NewTicker(retryWaiting)
	var err error
	for {
		if err = pg.setConnection(); err == nil {
			fmt.Println("connected to database successfully")
			return
		}
		// TODO: log error
		fmt.Println(err)
		<-tick.C
	}
}

func getDatabaseConfig(ctx context.Context) (config.ConnectionConfig, error) {
	env := environment.GetEnvironment()
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	conf, err := config.FromConfigFile(ctx, fmt.Sprintf("%v/data-access/postgres_service/config.json", path), env)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (pg *postgresServiceImpl) setConnection() error {
	if pg.engine != nil {
		return nil
	}
	ctx := context.Background()
	conf, err := getDatabaseConfig(ctx)
	if err != nil {
		return err
	}
	pg.conf = conf
	var pool *database.ConnectionPool
	pool, err = database.NewConnectionPool(pg.conf)
	if err != nil {
		return fmt.Errorf("failed establishing connection to database")
	}
	pg.engine = pool
	return nil
}

func (pg *postgresServiceImpl) SaveServiceLogs(logs []entities.ServiceLog) <-chan error {
	errors := make(chan error)
	go func() {
		defer close(errors)
		if err := pg.setConnection(); err != nil {
			errors <- err
			return
		}
		schemas := make([]relations.ServiceLogModel, 0, len(logs))
		for _, log := range logs {
			schemas = append(schemas, relations.NewServiceLogModelFromEntity(log))
		}
		pg.mu.Lock()
		defer pg.mu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := pg.engine.Conn.Begin(options)
		defer transaction.Commit()
		var err error
		chunkSize := 2000
		for i := 0; i < len(schemas); i += chunkSize {
			end := i + chunkSize
			if end >= len(schemas) {
				end = len(schemas)
			}
			records := schemas[i:end]
			chunkErr := pg.engine.Conn.Clauses(
				clause.OnConflict{
					Columns:   []clause.Column{},
					DoNothing: true,
				}).Create(&records).Error
			if chunkErr != nil {
				err = chunkErr
			}
		}
		errors <- err
	}()
	return errors
}
