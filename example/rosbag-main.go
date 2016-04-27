package main

import (
	"flag"

	"github.com/brychanrobot/goros"
)

func main() {
	flag.Parse()
	goros.ParseRosbag(flag.Arg(0))
}
