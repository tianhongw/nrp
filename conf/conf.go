package conf

import (
	"github.com/spf13/viper"
)

var gConfig *Config

func GetConfig() *Config {
	return gConfig
}

type Config struct {
	Server *ServerOption `mapstructure:"server"`
	Client *ClientOption `mapstructure:"client"`
	Log    *LogOption    `mapstructure:"log"`
}

type ServerOption struct {
	// public addr for HTTP connections
	HTTPAddr string `mapstructure:"http_addr"`

	// public addr for HTTPS connections
	HTTPSAddr string `mapstructure:"https_addr"`

	// public port listening for nrp client
	ClientAddr string `mapstructure:"client_addr"`

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

type ClientOption struct {
	ServerAddr string                   `mapstructure:"server_addr"`
	HTTPProxy  string                   `mapstructure:"http_proxy"`
	AuthToken  string                   `mapstructure:"auth_token"`
	Tunnels    map[string]*TunnelOption `mapstructure:"tunnels"`
}

type TunnelOption struct {
	HostName  string            `mapstructure:"host_name"`
	SubDomain string            `mapstructure:"sub_domain"`
	Protocols map[string]string `mapstructure:"protocols"`
	HttpAuth  string            `mapstructure:"http_auth"`
	// remote tcp port ask for
	RemotePort int `mapstructure:"remote_port"`
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
