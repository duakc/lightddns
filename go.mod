module github.com/duakc/lightddns

go 1.26.2

require (
	github.com/duakc/mt v0.0.0-20260518071304-19bf4e4b1c7a
	github.com/elastic/go-freelru v0.16.0
	github.com/goccy/go-yaml v1.19.2
	github.com/itchyny/gojq v0.12.19
	github.com/joho/godotenv v1.5.1
	github.com/miekg/dns v1.1.72
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	go.uber.org/zap v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.53.0
	golang.org/x/sys v0.43.0
	golang.org/x/tools v0.44.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.uber.org/zap => github.com/duakc/zap v0.0.0-20260409023011-7d65ec09a648
