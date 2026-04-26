package datasources

import "fmt"

type DatasourceNotFoundError struct {
	Name string
}

func (e *DatasourceNotFoundError) Error() string {
	return fmt.Sprintf("datasource `%s` not found", e.Name)
}

type EmptyGroupError struct {
	Type string
	Name string
}

func (e *EmptyGroupError) Error() string {
	return fmt.Sprintf("empty %s group: %s", e.Type, e.Name)
}
