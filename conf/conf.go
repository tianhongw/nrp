package conf

import (
	"sync"

	"github.com/spf13/viper"
)

var gConfig *Config
var mu sync.Mutex

func GetConfig() *Config {
	mu.Lock()
	defer mu.Unlock()

	return gConfig
}

type Config struct {
	Server *ServerOption `mapstructure:"server"`
	Log    *LogOption    `mapstructure:"log"`
}

type ServerOption struct {
	// public port for HTTP connections
	HTTPPort string `mapstructure:"http_addr"`

	// public port for HTTPS connections
	HTTPSPort string `mapstructure:"https_addr"`

	// public port listening for nrp client
	TunnelPort string `mapstructure:"tunnel_addr"`

	// domain where the tunnels are hosted
	Domain string `mapstructure:"domain"`

	// path to the tls certificate file
	TLSCrt string `mapstructure:"tls_crt"`

	// path to the tls key file
	TLSKey string `mapstructure:"tls_key"`

	// timeout in sec for connection write
	ConnWriteTimeoutSec int `mapstructure:"conn_write_timeout_sec"`

	// timeout in sec for connection read
	ConnReadTimeoutSec int `mapstructure:"conn_read_timeout_sec"`
}

type LogOption struct {
	Type         string   `mapstructure:"type"`
	Level        string   `mapstructure:"level"`
	Format       string   `mapstructure:"format"`
	Outputs      []string `mapstructure:"outputs"`
	ErrorOutputs []string `mapstructure:"error_outputs"`
}

func Init(cfgFile, cfgType string) (string, error) {
	v := viper.New()

	v.SetConfigFile(cfgFile)
	v.SetConfigType(cfgType)

	if err := v.ReadInConfig(); err != nil {
		return "", err
	}

	c := new(Config)

	if err := v.Unmarshal(c); err != nil {
		return "", err
	}

	gConfig = c

	return v.ConfigFileUsed(), nil
}
