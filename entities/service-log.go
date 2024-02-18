package entities

import (
	allowedLength "github.com/gerdooshell/tax-logger/entities/constants/allowed-length"
	"github.com/gerdooshell/tax-logger/entities/severity"
	"github.com/gerdooshell/tax-logger/lib/helper"
	"time"
)

type ServiceLog struct {
	Timestamp time.Time
	Severity  severity.Severity
	Message   string
	Origin    OriginLog
}

func (sl *ServiceLog) Validate() (err error) {
	if err = helper.ValidateLengthStr(sl.Message, allowedLength.MinLogMessageLength, allowedLength.MaxLogMessageLength); err != nil {
		return err
	}
	if err = helper.Sanitize(sl.Message); err != nil {
		return err
	}
	return sl.Origin.Validate()
}
