package proxy

import (
	"cmp"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Web      WebConfig       `yaml:"web"`
	Storage  StorageConfig   `yaml:"storage"`
	Backends []BackendConfig `yaml:"backends"`
}

type WebConfig struct {
	ExternalHost string `yaml:"externalHost"`
	ListenAddr   string `yaml:"listenAddr"`
	Port         uint16 `yaml:"port"`
	TLS          bool   `yaml:"tls"`
}

type StorageConfig struct {
	NZBDir string `yaml:"nzbDir"`
	DBPath string `yaml:"dbPath"`
}

type BackendConfig struct {
	Name    string     `yaml:"name"`
	BaseURL string     `yaml:"baseUrl"`
	APIKey  string     `yaml:"apiKey"`
	RSS     *RSSConfig `yaml:"rss,omitempty"`
}

type RSSConfig struct {
	RSSPath        string            `yaml:"rssPath"`
	RSSQueryParams map[string]string `yaml:"rssQueryParams"`
	Feeds          []RSSFeed         `yaml:"feeds"`
}

type RSSFeed struct {
	Name         string            `yaml:"name"`
	PollInterval time.Duration     `yaml:"pollInterval"`
	QueryParams  map[string]string `yaml:"queryParams"`
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
