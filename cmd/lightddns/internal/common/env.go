package common

import (
	"os"
	"path/filepath"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services/filehelper"

	"github.com/joho/godotenv"
)

// ApplyEnvFile loads KEY=VALUE pairs from envFile into the process environment
// via os.Setenv, so the whole program - including {{ .Env.KEY }} expansion in
// the config - sees them.
//
// Call it only when the user opts in with --env-file. Without it the inherited
// environment (e.g. systemd's EnvironmentFile=) is used unchanged: the daemon
// no longer silently picks up a stray .env.
func ApplyEnvFile(fileHelper filehelper.Helper, envFile string) error {
	data, err := readEnvFile(fileHelper, envFile)
	if err != nil {
		return err
	}
	parsed, err := godotenv.UnmarshalBytes(data)
	if err != nil {
		return err
	}
	for k, v := range parsed {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func readEnvFile(fileHelper filehelper.Helper, name string) ([]byte, error) {
	if filepath.IsAbs(name) {
		return os.ReadFile(name)
	}
	return fileHelper.Root().ReadFile(name)
}

func envMap() map[string]string {
	result := make(map[string]string)
	for _, kv := range os.Environ() {
		if k, v, ok := mt.KeyValue(kv); ok {
			result[k] = v
		}
	}
	return result
}
