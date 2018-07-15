package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

// BuildDate to record build date
var BuildDate = "Unknow Date"

// Version to record build bersion
var Version = "1.0.1"

const fileStoreName = "myshare-filestore"
const contractManager = "myshare-contract-manager"
const lsName = "lsdata"
const rsName = "rsdata"
const cmNum = 2

func printUsage(inspectCmd *flag.FlagSet, testCmd *flag.FlagSet) {
	flag.Usage()
	fmt.Println("  inspect")
	fmt.Println("        Inspect infomation")
	fmt.Println("  test")
	fmt.Println("        Test command")
	inspectCmd.Usage()
	testCmd.Usage()
}

// Contract struct
type Contract struct {
	Version        int    `json:"version"`
	Fiber          string `json:"fiber"`
	Miner          string `json:"miner"`
	MinerFootprint string `json:"minerFootprint"`
	Hash           string `json:"hash"`
	Size           int    `json:"size"`
	LeaseBegin     bool   `json:"leaseBegin"`
	LeaseEnd       bool   `json:"leaseEnd"`
	Status         string `json:"status"`
}

func internalKey(key string) string {
	hash := sha1.New()
	hash.Write([]byte(key))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getChunkKeyByIndex(key string, index int) string {
	return fmt.Sprintf("%s-%06d", key, index)
}

func printContract(contract *Contract) {
	fmt.Printf("%v", contract)
}

func inspect(root *string, key *string, sizeThreshold *int) {
	fileStoreRoot := filepath.Join(*root, fileStoreName)
	cmRoot := filepath.Join(*root, contractManager)
	lsRoot := filepath.Join(fileStoreRoot, lsName)
	rsRoot := filepath.Join(fileStoreRoot, rsName)
	inKey := internalKey(*key)

	cmDb, err := leveldb.OpenFile(cmRoot, nil)
	if err != nil {
		fmt.Println("Open leveldb error: ", err)
		return
	}
	defer cmDb.Close()

	cdata, err := cmDb.Get([]byte(inKey), nil)
	if err != nil {
		fmt.Println("not find contract", err)
	}

	contract := Contract{}
	if err := json.Unmarshal(cdata, &contract); err != nil {
		fmt.Println("decode contract error: ", err)
		return
	}

	printContract(&contract)
	if contract.Status != "MINER_USED" {
		fmt.Println("Invalid constract status: ", contract.Status)
		return
	}

	if contract.Size > *sizeThreshold {
		/**
		 * key : 12345678985a0aa21c23f5abd2975a89b682abcd
		 * path: 123/456/789/85a0aa21c23f5abd2975a89b682abcd
		 */

		filename := filepath.Join(rsRoot, inKey[0:2], inKey[3:5], inKey[6:8], inKey[9:])
		fd, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer fd.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, fd); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Data information:")
		fmt.Println("  store       : RS")
		fmt.Println("  internal Key: ", inKey)
		fmt.Println("  path        : ", filename)
		fmt.Printf("   sha256 hash  : %x\n", hash.Sum(nil))
	} else {
		// getDB index get DB from key
		dbIndex := int(inKey[0]) % cmNum
		fmt.Println("lsDB key index", inKey, dbIndex)
		dbPath := filepath.Join(lsRoot, strconv.Itoa(dbIndex))
		fmt.Println("Real db path: ", dbPath)

		lsDb, err := leveldb.OpenFile(cmRoot, nil)
		if err != nil {
			fmt.Println("Open leveldb error: ", err)
			return
		}
		defer lsDb.Close()

		chunkIndex := 0
		hash := sha256.New()
		for {
			chunkKey := getChunkKeyByIndex(inKey, chunkIndex)
			ldata, err := lsDb.Get([]byte(chunkKey), nil)
			if err != nil {
				fmt.Println("not find block in db", err)
				break
			}
			hash.Write(ldata)
			chunkIndex++
		}

		fmt.Println("Data information:")
		fmt.Println("  store        : LS")
		fmt.Println("  internal Key : ", inKey)
		fmt.Println("  CM index     : ", dbIndex)
		fmt.Println("  Chunk number : ", chunkIndex)
		fmt.Printf("   sha256 hash  : %x\n", hash.Sum(nil))
	}

}

func main() {
	var VERSION = fmt.Sprintf("Version: %s  build: %s", Version, BuildDate)
	version := flag.Bool("version", false, "Output version")
	help := flag.Bool("help", false, "Output help page")

	inspectCmd := flag.NewFlagSet("inspect", flag.ExitOnError)
	inspectKey := inspectCmd.String("key", "", "Block key")
	inspectRoot := inspectCmd.String("root", "", "Data root dir")
	inspectSize := inspectCmd.Int("size", 102400, "Block size threshold")

	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	testApp := testCmd.String("app", "", "test command")

	if len(os.Args) < 2 {
		printUsage(inspectCmd, testCmd)
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
		fmt.Printf("You asked: %q  %q %q\n", *inspectKey, *inspectRoot, *inspectSize)
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
			printUsage(inspectCmd, testCmd)
		} else {
			fmt.Println("Unknow args ...")
			printUsage(inspectCmd, testCmd)
		}
	}
}
