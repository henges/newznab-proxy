package proxy

import (
	"cmp"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Backends []struct {
		Name    string `yaml:"name"`
		BaseURL string `yaml:"baseUrl"`
		APIKey  string `yaml:"apiKey"`
	} `yaml:"backends"`
}

const configPathEnvVar = "NEWZNAB_PROXY_CONFIG_PATH"

func MustGetConfig() *Config {

	confPath := cmp.Or(os.Getenv(configPathEnvVar), ".my.config.yaml")
	f, err := os.Open(confPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var c Config
	err = yaml.NewDecoder(f).Decode(&c)
	if err != nil {
		panic(err)
	}
	return &c
}
