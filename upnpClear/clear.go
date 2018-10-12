package main

import (
	"fmt"
	"github.com/NebulousLabs/go-upnp"
)

func main() {
	var maxPort uint16 = 30000
	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		fmt.Println("Discover error: ", err)
		return
	}

	for i := uint16(20000); i < maxPort; i++ {
		// un-forward a port
		err = d.Clear(i)
		if err == nil {
		       fmt.Printf("unmap port:%d success\n", i)
		}
	}

}
