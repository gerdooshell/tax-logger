package entities

import (
	allowedLength "github.com/gerdooshell/tax-logger/entities/constants/allowed-length"
	serviceName "github.com/gerdooshell/tax-logger/entities/constants/service-name"
	"github.com/gerdooshell/tax-logger/lib/helper"
)

type OriginLog struct {
	ServiceName serviceName.ServiceName
	StackTrace  string
	ProcessId   string
}

func (o *OriginLog) Validate() (err error) {
	if err = helper.ValidateLengthStr(o.StackTrace, allowedLength.MinStackTraceLength, allowedLength.MaxStackTraceLength); err != nil {
		return err
	}
	if err = helper.ValidateLengthStr(o.ProcessId, allowedLength.MinProcessIdLength, allowedLength.MaxProcessIdLength); err != nil {
		return err
	}
	texts := []string{o.StackTrace, o.ProcessId}
	if err = helper.SanitizeAll(texts); err != nil {
		return err
	}
	return nil
}
