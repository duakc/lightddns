package gendoc

type LangCode string

const (
	LangDefault = LangEnUS

	LangZhCn LangCode = "ZH_CN"
	LangEnUS LangCode = "EN_US"
)

var LangList = []LangCode{LangZhCn, LangEnUS}

var LangMap map[string]LangCode

func init() {
	LangMap = make(map[string]LangCode)
	for _, v := range LangList {
		LangMap[string(v)] = v
	}
}
