package badyaml

import (
	"fmt"
	"regexp"
	"time"

	"github.com/duakc/mt"

	"github.com/itchyny/gojq"
	"go.uber.org/zap/zapcore"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
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

type LogLevel zapcore.Level

func (L *LogLevel) UnmarshalYAML(data []byte) error {
	s, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	level, err := zapcore.ParseLevel(s)
	if err != nil {
		return err
	}
	*L = LogLevel(level)
	return nil
}

//type NotEmpty[T any] struct {
//	Value T
//}
//
//var ErrValuesIsEmpty = errors.New("values is empty")
//
//func (E *NotEmpty[T]) UnmarshalYAML(data []byte) error {
//	ret, err := UnmarshalType[T](data)
//	if err != nil {
//		return err
//	}
//
//	type ZeroChecker interface {
//		IsZero() bool
//	}
//
//	if checker, ok := any(ret).(ZeroChecker); ok && checker.IsZero() {
//		return ErrValuesIsEmpty
//	}
//
//	val := reflect.ValueOf(ret)
//
//	if !val.IsValid() {
//		return fmt.Errorf("value is invalid")
//	}
//
//	kind := val.Kind()
//
//	switch {
//	case (kind == reflect.Slice || kind == reflect.Map) && val.Len() == 0:
//		return ErrValuesIsEmpty
//	case (kind == reflect.Ptr || kind == reflect.Interface || kind == reflect.Chan ||
//		kind == reflect.Func) && val.IsNil():
//		return ErrValuesIsEmpty
//	case val.IsZero():
//		return ErrValuesIsEmpty
//	}
//
//	E.Value = ret
//
//	return nil
//}

type JQ gojq.Query

func (jq *JQ) UnmarshalYAML(data []byte) error {
	jqString, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	jqQuery, jqParseErr := gojq.Parse(jqString)
	if jqParseErr != nil {
		return jqParseErr
	}

	*jq = JQ(*jqQuery)
	return nil
}

func (jq *JQ) Cast() *gojq.Query {
	if jq == nil {
		return nil
	}
	return (*gojq.Query)(jq)
}

type Regex regexp.Regexp

func (re *Regex) UnmarshalYAML(data []byte) error {
	reString, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	compile, compileErr := regexp.Compile(reString)
	if compileErr != nil {
		return compileErr
	}

	*re = Regex(*compile)
	return nil
}

func (re *Regex) Cast() *regexp.Regexp {
	if re == nil {
		return nil
	}
	return (*regexp.Regexp)(re)
}
