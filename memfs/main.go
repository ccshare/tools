package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"memfs/fs"

	"github.com/jacobsa/fuse"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		fmt.Println(flag.Args())
		return
	}

	fs := memfs.NewMemFS(0, 0)

	fmt.Printf("create fs Pid[%v]\n", os.Getpid())

	cfg := &fuse.MountConfig{
		FSName: "memfs",
	}
	mfs, err := fuse.Mount(flag.Arg(0), fs, cfg)
	if err != nil {
		fmt.Println("mount error: ", err)
	}

	fmt.Printf("success mount, mountpoint[%v], Pid[%v]\n", mfs.Dir(), os.Getpid())

	mfs.Join(context.Background())
}
