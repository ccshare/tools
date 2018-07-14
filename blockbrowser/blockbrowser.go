package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// BuildDate to record build date
var BuildDate = "Unknow Date"

// Version to record build bersion
var Version = "1.0.1"

func printUsage() {
	fmt.Printf("usage: %s [args] <command> [<args>]\n", os.Args[0])
	fmt.Printf("%s args:\n", os.Args[0])
	fmt.Println("  -help     Print this help page")
	fmt.Println("  -version  Print version")
	fmt.Println("commands: ")
	fmt.Println("  inspect   Inspcet info")
	fmt.Println("     -key   Object key")
	fmt.Println("     -root  Data root dir")
	fmt.Println("  test      Test command")
	fmt.Println("     -app      Test command name")
}

func main() {
	var VERSION = fmt.Sprintf("Version: %s  build: %s", Version, BuildDate)
	version := flag.Bool("version", false, "Output version")
	help := flag.Bool("help", false, "Output help page")

	inspectCmd := flag.NewFlagSet("inspect", flag.ExitOnError)
	inspectKey := inspectCmd.String("key", "", "file key")
	inspectRoot := inspectCmd.String("root", "", "data root dir")

	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	testApp := testCmd.String("app", "", "test command")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "inspect":
		inspectCmd.Parse(os.Args[2:])
	case "test":
		testCmd.Parse(os.Args[2:])
	default:
		flag.Parse()
	}

	if inspectCmd.Parsed() {
		if *inspectKey == "" {
			fmt.Println("Please supply the key using -key option.")
			os.Exit(2)
		} else if *inspectRoot == "" {
			fmt.Println("Please supply the root using -root option.")
			os.Exit(3)
		}
		fmt.Printf("You asked: %q  %q\n", *inspectKey, *inspectRoot)
	} else if testCmd.Parsed() {
		if *testApp == "" {
			fmt.Println("Please supply the user using -user option.")
			os.Exit(2)
		}
		fmt.Printf("You asked: %q\n", *testApp)
	} else { // if flag.Parsed()
		if true == *version {
			fmt.Printf("%s  %s\n", filepath.Base(os.Args[0]), VERSION)
		} else if true == *help {
			printUsage()
		} else {
			fmt.Println("Unknow args ...")
			printUsage()
		}
	}
}
