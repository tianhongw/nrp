package conf

// httpAddr := flag.String("httpAddr", ":80", "Public address for HTTP connections, empty string to disable")
// httpsAddr := flag.String("httpsAddr", ":443", "Public address listening for HTTPS connections, emptry string to disable")
// tunnelAddr := flag.String("tunnelAddr", ":4443", "Public address listening for ngrok client")
// domain := flag.String("domain", "ngrok.com", "Domain where the tunnels are hosted")
// tlsCrt := flag.String("tlsCrt", "", "Path to a TLS certificate file")
// tlsKey := flag.String("tlsKey", "", "Path to a TLS key file")

type Config struct {
	Server *Server `mapstructure:"server"`
}

type Server struct {
	// public address for HTTP connections
	HTTPAddr string `mapstructure:"http_addr"`

	// public address for HTTPS connections
	HTTPSAddr string `mapstructure:"https_addr"`

	// public address listening for nrp client
	TunnelAddr string `mapstructure:"tunnel_addr"`

	// domain where the tunnels are hosted
	Domain string `mapstructure:"domain"`

	// path to the tls certificate file
	TLSCrt string `mapstructure:"tls_crt"`

	// path to the tls key file
	TLSKey string `mapstructure:"tls_key"`
}
