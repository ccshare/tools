package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/diskv"
)

var (
	cachePath string
)

func main() {
	flag.StringVar(&cachePath, "d", "cache", "cache dir")
	flag.Parse()

	// Simplest transform function: put all the data files into the base dir.
	flatTransform := func(s string) []string {
		ret := strings.Split(s, "/")
		fmt.Println("transform: ", s, "->", ret)
		return ret
	}

	// Initialize a new diskv store
	d := diskv.New(diskv.Options{
		BasePath:     cachePath,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})

	// Write three bytes to the key "alpha".
	key := "a/b/c/d"
	fmt.Println("begein Write")
	err := d.Write(key, []byte("value-of-key"))
	if err != nil {
		fmt.Println("Write error: ", err)
	} else {
		fmt.Println("Write success")
	}
	// Read the value back out of the store.
	fmt.Println("begein Read")
	value, err := d.Read(key)
	if err != nil {
		fmt.Println("read error: ", err)
	} else {
		fmt.Printf("%s\n", value)
	}

	// Erase the key+value from the store (and the disk).
	//d.Erase(key)

}
