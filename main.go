package main

import (
	"log"

	"github.com/tianhongw/grp/cmd"
)

func main() {
	if err := cmd.NewCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
