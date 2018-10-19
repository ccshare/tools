package main

import (
	"fmt"
	"math"
	"flag"
	"github.com/NebulousLabs/go-upnp"
)

func main() {
	port := flag.Uint("port", 20000, "port")
	max := flag.Uint("max", 30000, "max port")
	flag.Parse()
	if *max < *port || *port >math.MaxUint16 || *max > math.MaxUint16 {
		fmt.Println("invalid port")
		return
	}
	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		fmt.Println("Discover error: ", err)
		return
	}

	for i := *port; i < *max; i++ {
		iPort := uint16(i)
		// un-forward a port
		enabled, err := d.IsForwardedTCP(iPort)
		if err != nil {
			continue
		}
		if enabled == true {
			err = d.Clear(iPort)
			if err == nil {
				fmt.Printf("unmap port:%d success\n", iPort)
			}
		}
	}

}
