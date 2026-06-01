package ipserver

//go:generate go run gen_serializers.go

type Response struct {
	IP      string `json:"ip"       yaml:"ip"`
	IsBogon bool   `json:"is_bogon" yaml:"is_bogon"`
	Time    string `json:"time"     yaml:"time"`
}
