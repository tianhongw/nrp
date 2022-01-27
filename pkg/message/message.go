package message

import "reflect"

var TypeMap map[string]reflect.Type

func init() {
	TypeMap = make(map[string]reflect.Type)

	TypeMap["AuthRequest"] = toReflectType((*AuthRequest)(nil))
	TypeMap["AuthResponse"] = toReflectType((*AuthResponse)(nil))
	TypeMap["TunnelRequest"] = toReflectType((*TunnelRequest)(nil))
	TypeMap["TunnelResponse"] = toReflectType((*TunnelResponse)(nil))
	TypeMap["ProxyRequest"] = toReflectType((*ProxyRequest)(nil))
	TypeMap["ProxyReg"] = toReflectType((*ProxyReg)(nil))
	TypeMap["ProxyStart"] = toReflectType((*ProxyStart)(nil))
	TypeMap["Ping"] = toReflectType((*Ping)(nil))
	TypeMap["Pong"] = toReflectType((*Pong)(nil))
}

func toReflectType(v interface{}) reflect.Type {
	return reflect.TypeOf(v).Elem()
}

type Message interface{}

// client to server
type AuthRequest struct {
	User     string
	Password string
	ClientId string
}

// server to client
type AuthResponse struct {
	ClientId string
	ErrorMsg string
}

// client to server
type TunnelRequest struct {
	RequestId string
	Protocal  string

	// http only
	HostName  string
	SubDomain string
	HttpAuth  string

	// tcp only
	RemotePort uint16
}

// server to client
type TunnelResponse struct {
	RequestId string
	URL       string
	Protocol  string
	ErrorMsg  string
}

// server to client
type ProxyRequest struct {
}

// client to server
type ProxyReg struct {
	ClientId string
}

// server to client
type ProxyStart struct {
	URL        string
	ClientAddr string
}

// client to server or server to client
type Ping struct {
}

// client to server or server to client
type Pong struct {
}
