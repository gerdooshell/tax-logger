package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gerdooshell/tax-logger/environment"
	dbEngine "github.com/gerdooshell/tax-logger/lib/database/engine"
	dbHosting "github.com/gerdooshell/tax-logger/lib/database/hosting"
	"github.com/gerdooshell/tax-logger/secret-service/azure"
)

type connectionConfig struct {
	Engine          dbEngine.Engine   `json:"Engine"`
	HostTag         dbHosting.HostTag `json:"HostTag"`
	ContainsSecrets bool              `json:"ContainsSecrets"`
	VaultURL        string            `json:"VaultURL"`
	Host            string            `json:"Host"`
	User            string            `json:"User"`
	Password        string            `json:"Password"`
	Port            string            `json:"Port"`
	Database        string            `json:"Database"`
	SSL             bool              `json:"SSL"`
}

func FromConfigFile(ctx context.Context, absFilePath string, env environment.Environment) (ConnectionConfig, error) {
	data, err := os.ReadFile(absFilePath)
	if err != nil {
		return nil, err
	}
	var confMap map[environment.Environment]connectionConfig
	if err = json.Unmarshal(data, &confMap); err != nil {
		return nil, err
	}
	conf, ok := confMap[env]
	if !ok {
		return nil, errors.New(fmt.Sprintf("no config found for environment %v", env))
	}
	if conf.ContainsSecrets {
		conf.VaultURL = strings.Trim(conf.VaultURL, " ")
		if conf.VaultURL == "" {
			return nil, errors.New("invalid vault url")
		}
		if conf, err = setVaultSecrets(ctx, conf); err != nil {
			return nil, err
		}
	}
	return &conf, nil
}

func setVaultSecrets(ctx context.Context, conf connectionConfig) (connectionConfig, error) {
	timeout := time.Second * 5
	azVault := azure.NewSecretService(conf.VaultURL)
	host, errHost := azVault.GetSecretValue(ctx, conf.Host)
	user, errUser := azVault.GetSecretValue(ctx, conf.User)
	pass, errPass := azVault.GetSecretValue(ctx, conf.Password)
	port, errPort := azVault.GetSecretValue(ctx, conf.Port)
	db, errDb := azVault.GetSecretValue(ctx, conf.Database)
	select {
	case conf.Host = <-host:
	case err := <-errHost:
		return conf, err
	case <-time.After(timeout):
		return conf, fmt.Errorf("fetching host secret timed out")
	}
	select {
	case conf.User = <-user:
	case err := <-errUser:
		return conf, err
	case <-time.After(timeout):
		return conf, fmt.Errorf("fetching user secret timed out")
	}
	select {
	case conf.Password = <-pass:
	case err := <-errPass:
		return conf, err
	case <-time.After(timeout):
		return conf, fmt.Errorf("fetching password secret timed out")
	}
	select {
	case conf.Port = <-port:
	case err := <-errPort:
		return conf, err
	case <-time.After(timeout):
		return conf, fmt.Errorf("fetching port secret timed out")
	}
	select {
	case conf.Database = <-db:
	case err := <-errDb:
		return conf, err
	}
	return conf, nil
}

func (c *connectionConfig) GetConnectionString() string {
	if c.Engine == dbEngine.Postgres {
		return c.toPostgresConnectionString()
	}
	return ""
}

func (c *connectionConfig) toPostgresConnectionString() string {
	sslMode := "disable"
	if c.SSL {
		sslMode = "enable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, sslMode)
}

func (c *connectionConfig) GetSignature() struct {
	Engine  dbEngine.Engine
	HostTag dbHosting.HostTag
} {
	return struct {
		Engine  dbEngine.Engine
		HostTag dbHosting.HostTag
	}{Engine: c.Engine, HostTag: c.HostTag}
}

func (c *connectionConfig) GetEngine() dbEngine.Engine {
	return c.Engine
}

func (c *connectionConfig) GetHostTag() dbHosting.HostTag {
	return c.HostTag
}
