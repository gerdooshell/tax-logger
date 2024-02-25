package serviceName

import "fmt"

type ServiceName string

const (
	NotFound     ServiceName = "not-found"
	TaxCore      ServiceName = "tax-core"
	DataProvider ServiceName = "data-provider"
	Mock         ServiceName = "mock"
)

func (s ServiceName) ToString() string {
	return string(s)
}

func FromString(serviceNameStr string) (ServiceName, error) {
	var err error
	var serviceName ServiceName
	switch serviceNameStr {
	case TaxCore.ToString():
		serviceName = TaxCore
	case DataProvider.ToString():
		serviceName = DataProvider
	case Mock.ToString():
		serviceName = Mock
	default:
		serviceName = NotFound
		err = fmt.Errorf("invalid service name: \"%s\"", serviceNameStr)
	}
	return serviceName, err
}
