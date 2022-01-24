package message

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Envelope struct {
	Type    string
	Payload json.RawMessage
}

func Unpack(buf []byte) (Message, error) {
	var env Envelope
	if err := json.Unmarshal(buf, &env); err != nil {
		return nil, err
	}

	t, ok := TypeMap[env.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported message type: %s", env.Type)
	}

	msg := reflect.New(t).Interface().(Message)

	if err := json.Unmarshal(env.Payload, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func Pack(payload interface{}) ([]byte, error) {
	return json.Marshal(struct {
		Type    string
		Payload interface{}
	}{
		Type:    reflect.TypeOf(payload).Elem().Name(),
		Payload: payload})
}
