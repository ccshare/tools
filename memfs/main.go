package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jacobsa/fuse"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		fmt.Println(flag.Args())
		return
	}

	fs := NewMemFS(0, 0)
	logger := log.New(os.Stderr, "", log.LstdFlags)

	logger.Printf("create fs Pid[%v]\n", os.Getpid())

	cfg := &fuse.MountConfig{
		FSName:      "memfs",
		ErrorLogger: logger,
		DebugLogger: logger,
	}
	mfs, err := fuse.Mount(flag.Arg(0), fs, cfg)
	if err != nil {
		fmt.Println("mount error: ", err)
	}

	logger.Printf("success mount, mountpoint[%v], Pid[%v]\n", mfs.Dir(), os.Getpid())

	mfs.Join(context.Background())
}
