package datasourcex

import (
	"fmt"

	"github.com/duakc/lightddns/adapter"
)

type DatasourceNotFoundError struct {
	Err error
}

func (e *DatasourceNotFoundError) Error() string {
	return fmt.Sprintf("datasource not found: %s", e.Err.Error())
}

func (e *DatasourceNotFoundError) Unwrap() error {
	return e.Err
}

type DatasourceError struct {
	Err       error
	IPVersion string
	Name      string
	Type      string
}

func newDatasourceError(err error, ipVersion string, ds adapter.Datasource) *DatasourceError {
	return &DatasourceError{
		Err:       err,
		IPVersion: ipVersion,
		Name:      ds.Name(),
		Type:      ds.Type(),
	}
}

func (e *DatasourceError) Error() string {
	return fmt.Sprintf("get ipv%s addresses from datasource(%s,%s) failed: %s",
		e.IPVersion, e.Type, e.Name, e.Err.Error())
}

func (e *DatasourceError) Unwrap() error {
	return e.Err
}
