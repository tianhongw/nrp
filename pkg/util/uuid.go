package util

import "github.com/bwmarrin/snowflake"

var gsf *snowflake.Node

func init() {
	if n, err := snowflake.NewNode(1); err != nil {
		panic(err)
	} else {
		gsf = n
	}
}

func NewIntID() int64 {
	return gsf.Generate().Int64()
}

func NewStringID() string {
	return gsf.Generate().String()
}
