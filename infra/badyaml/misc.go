package badyaml

import (
	"fmt"
	"time"

	"github.com/duakc/mt"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(data []byte) error {
	s := mt.UnquoteString(string(data))
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

type EnvironmentVariable struct {
	indexMap map[string]string
	Values   []string
}

func (E *EnvironmentVariable) UnmarshalYAML(data []byte) error {
	l := Listable[string]{}
	err := Unmarshal(data, &l)
	if err != nil {
		return err
	}
	E.indexMap = make(map[string]string, len(l.Value))
	for i := 0; i < len(l.Value); i++ {
		rawEnv := l.Value[i]
		k, v, found := mt.KeyValue(rawEnv)
		if !found {
			return fmt.Errorf("invalid environment variable: %s", rawEnv)
		}
		E.indexMap[k] = v
	}
	E.Values = l.Value
	return nil
}

func (E *EnvironmentVariable) Lookup(key string) string {
	vv, ok := E.indexMap[key]
	if ok {
		return vv
	}
	return ""
}
