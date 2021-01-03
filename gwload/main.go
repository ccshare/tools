package main

import (
	"flag"
	"fmt"
)

var (
	gw string
	sc string
)

func main() {
	flag.StringVar(&gw, "gw", "", "gw address")
	flag.StringVar(&sc, "sc", "", "sc address")
	flag.Parse()
	fmt.Println("load")
}
