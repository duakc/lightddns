package gendoc

type LangCode string

var RegisteredOption = []struct {
	Name       string
	MappedName map[LangCode]string
}{
	{
		Name: "LogOption",
		MappedName: map[LangCode]string{
			LangZhCn: "日志选项",
			LangEnUS: "Log Option",
		},
	},
}

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
