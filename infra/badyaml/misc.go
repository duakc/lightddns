package badyaml

import (
	"time"

	"github.com/duakc/lightddns/infra/common"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(data []byte) error {
	s := common.UnquoteString(string(data))
	if len(s) == 0 {
		*d = Duration(0)
		return nil
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}
