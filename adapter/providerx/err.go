package providerx

import (
	"fmt"
	"time"

	constpkg "github.com/duakc/lightddns/constant"
)

type ProviderNotFoundError struct {
	Err error
}

func (e *ProviderNotFoundError) Error() string {
	return fmt.Sprintf("provider not found: %s", e.Err.Error())
}

func (e *ProviderNotFoundError) Unwrap() error {
	return e.Err
}

func UpdateMessage(old string) string {
	if len(old) != 0 {
		return fmt.Sprintf("%s: last update %s, last record: %s",
			constpkg.Project, time.Now().Format(time.RFC3339), old)
	}

	return fmt.Sprintf("%s: last update %s",
		constpkg.Project, time.Now().Format(time.RFC3339))
}
