package severity

import "fmt"

type Severity string

const (
	Debug   Severity = "debug"
	Info    Severity = "info"
	Warning Severity = "warning"
	Error   Severity = "error"
	Fatal   Severity = "fatal"
)

func (s Severity) ToString() string {
	return string(s)
}

func FromString(severityStr string) (severity Severity, err error) {
	switch severityStr {
	case Debug.ToString():
		severity = Debug
	case Info.ToString():
		severity = Info
	case Warning.ToString():
		severity = Warning
	case Error.ToString():
		severity = Error
	case Fatal.ToString():
		severity = Fatal
	default:
		severity = Debug
		err = fmt.Errorf("invalid severity \"%s\"", severityStr)
	}
	return
}
