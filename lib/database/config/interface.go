package config

import (
	dbEngine "github.com/gerdooshell/tax-logger/lib/database/engine"
	dbHosting "github.com/gerdooshell/tax-logger/lib/database/hosting"
)

type ConnectionConfig interface {
	GetConnectionString() string
	GetSignature() struct {
		Engine  dbEngine.Engine
		HostTag dbHosting.HostTag
	}
	GetEngine() dbEngine.Engine
	GetHostTag() dbHosting.HostTag
}
