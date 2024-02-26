package interactors

import "github.com/gerdooshell/tax-logger/entities"

type DataAccess interface {
	SaveServiceLogs(logs []entities.ServiceLog) <-chan error
}
