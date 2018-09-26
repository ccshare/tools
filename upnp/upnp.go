package main

import (
	"fmt"
	"github.com/NebulousLabs/go-upnp"
	"log"
)

func main() {
	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		log.Fatal(err)
	}

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Your external IP is:", ip)

	// forward a port
	err = d.Forward(9001, "upnp-checker")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("port map success")

	// un-forward a port
	err = d.Clear(9001)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("port unmap success")

	// record router's location
	loc := d.Location()
	fmt.Println("router upnp info: ", loc)

	// connect to router directly
	d, err = upnp.Load(loc)
	if err != nil {
		log.Fatal(err)
	}
}
