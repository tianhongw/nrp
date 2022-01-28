package message

import (
	"encoding/binary"
	"fmt"

	"github.com/tianhongw/grp/pkg/conn"
)

func ReadMsg(c conn.IConn) (Message, error) {
	var size int64
	if err := binary.Read(c, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	buf := make([]byte, size)

	n, err := c.Read(buf)
	if err != nil {
		return nil, err
	}

	if int64(n) != size {
		return nil, fmt.Errorf("expected: %d bytes, but got: %d bytes", size, n)
	}

	return Unpack(buf)
}

func WriteMsg(c conn.IConn, msg Message) error {
	buf, err := Pack(msg)
	if err != nil {
		return err
	}

	if err := binary.Write(c, binary.LittleEndian, int64(len(buf))); err != nil {
		return err
	}

	_, err = c.Write(buf)

	return err
}
