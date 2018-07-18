package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

func printUsage(inspectCmd *flag.FlagSet, keyCmd *flag.FlagSet) {
	flag.Usage()
	fmt.Println("  inspect")
	fmt.Println("        Inspect infomation")
	fmt.Println("  key")
	fmt.Println("        Key convert")
	inspectCmd.Usage()
	keyCmd.Usage()
}

// Contract struct
type Contract struct {
	Version        int         `json:"version"`
	Fiber          string      `json:"fiber"`
	Miner          string      `json:"miner"`
	MinerFootprint string      `json:"minerFootprint"`
	Hash           string      `json:"hash"`
	Size           int         `json:"size"`
	LeaseBegin     json.Number `json:"leaseBegin"`
	LeaseEnd       json.Number `json:"leaseEnd"`
	Status         string      `json:"status"`
}

func internalKey(key string) string {
	hash := sha1.New()
	hash.Write([]byte(key))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getChunkKeyByIndex(key string, index int) string {
	return fmt.Sprintf("%s-%06d", key, index)
}

func getCmIndexFromKey(key string, cm int) (int, error) {
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return 0, err
	}
	return int(decoded[0]) % cm, nil
}

func inspect(root *string, key *string, sizeThreshold int, cmNum int) {
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

	cdata, err := cmDb.Get([]byte(*key), nil)
	if err != nil {
		fmt.Println("not find contract", *key, err)
		return
	}

	contract := Contract{}
	if err := json.Unmarshal(cdata, &contract); err != nil {
		fmt.Println("decode contract error: ", err)
		return
	}

	leaseBegin, _ := contract.LeaseBegin.Int64()
	leaseEnd, _ := contract.LeaseEnd.Int64()

	fmt.Printf("Contract information:\n")
	fmt.Printf("  hash      : %s\n", contract.Hash)
	fmt.Printf("  size      : %d (%dk)\n", contract.Size, contract.Size/1024)
	fmt.Printf("  leaseBegin: %s (%v)\n", contract.LeaseBegin, time.Unix(leaseBegin/1000, leaseBegin%1000))
	fmt.Printf("  leaseEnd  : %s (%v)\n", contract.LeaseEnd, time.Unix(leaseEnd/1000, leaseEnd%1000))
	fmt.Printf("  status    : %s\n", contract.Status)
	/*
		if contract.Status != "MINER_USED" {
			fmt.Println("Invalid constract status: ", contract.Status)
			return
		}
	*/

	if contract.Size > sizeThreshold {
		/**
		 * key : 12345678985a0aa21c23f5abd2975a89b682abcd
		 * path: 123/456/78985a0aa21c23f5abd2975a89b682abcd
		 */
		filename := filepath.Join(rsRoot, inKey[0:3], inKey[3:6], inKey[6:])
		fd, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer fd.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, fd); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Block information:\n")
		fmt.Printf("  store     : RS\n")
		fmt.Printf("  inKey     : %s\n", inKey)
		fmt.Printf("  path      : %s\n", filename)
		fmt.Printf("  hash      : %x\n", hash.Sum(nil))
	} else {
		// getDB index get DB from key
		dbIndex, err := getCmIndexFromKey(inKey, cmNum)
		if err != nil {
			log.Println(err)
		}
		dbPath := filepath.Join(lsRoot, strconv.Itoa(dbIndex))
		lsDb, err := leveldb.OpenFile(dbPath, nil)
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
				break
			}
			hash.Write(ldata)
			chunkIndex++
		}

		if chunkIndex == 0 {
			fmt.Printf("Block information(not find)\n")
		} else {
			fmt.Printf("Block information:\n")
			fmt.Printf("  store     : LS\n")
			fmt.Printf("  inKey     : %s\n", inKey)
			fmt.Printf("  CM Index  : %d\n", dbIndex)
			fmt.Printf("  Chunk NUM : %d\n", chunkIndex)
			fmt.Printf("  hash      : %x\n", hash.Sum(nil))
		}
	}

}

func main() {
	var VERSION = fmt.Sprintf("Version: %s  build: %s", Version, BuildDate)
	version := flag.Bool("v", false, "Output version")
	help := flag.Bool("h", false, "Output help page")

	inspectCmd := flag.NewFlagSet("inspect", flag.ExitOnError)
	inspectKey := inspectCmd.String("k", "", "Block key")
	inspectRoot := inspectCmd.String("r", "", "Data root dir")
	inspectSize := inspectCmd.Int("s", 102400, "Block size threshold")
	inspectCmnum := inspectCmd.Int("n", 2, "CM number")

	keyCmd := flag.NewFlagSet("key", flag.ExitOnError)
	keyKey := keyCmd.String("k", "", "block key")

	if len(os.Args) < 2 {
		printUsage(inspectCmd, keyCmd)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "inspect":
		inspectCmd.Parse(os.Args[2:])
	case "key":
		keyCmd.Parse(os.Args[2:])
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
		inspect(inspectRoot, inspectKey, *inspectSize, *inspectCmnum)
	} else if keyCmd.Parsed() {
		if *keyKey == "" {
			fmt.Println("Please supply the user using -user option.")
			os.Exit(2)
		}
		fmt.Printf("internalKey: %s\n", internalKey(*keyKey))
	} else { // if flag.Parsed()
		if true == *version {
			fmt.Printf("%s  %s\n", filepath.Base(os.Args[0]), VERSION)
		} else if true == *help {
			printUsage(inspectCmd, keyCmd)
		} else {
			fmt.Println("Unknow args ...")
			printUsage(inspectCmd, keyCmd)
		}
	}
}
