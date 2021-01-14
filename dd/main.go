package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	filename string
	offset   int64
	content  string
)

func main() {
	flag.StringVar(&filename, "f", "", "filename")
	flag.Int64Var(&offset, "p", 0, "offset")
	flag.StringVar(&content, "s", fmt.Sprintf("time: %s", time.Now().String()), "content")
	flag.Parse()

	if filename == "" {
		flag.Usage()
		return
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open error: ", err)
		return
	}
	defer f.Close()

	_, err = f.WriteAt([]byte(content), offset)
	if err != nil {
		fmt.Println("writeAt error: ", err)
	}

}
