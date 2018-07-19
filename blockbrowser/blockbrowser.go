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

func printUsage(inspectCmd *flag.FlagSet, fileCmd *flag.FlagSet, dbCmd *flag.FlagSet) {
	flag.Usage()
	fmt.Println("  inspect")
	fmt.Println("        Inspect infomation")
	fmt.Println("  file")
	fmt.Println("        Read filesystem")
	fmt.Println("  db")
	fmt.Println("        Read DB")
	inspectCmd.Usage()
	fileCmd.Usage()
	dbCmd.Usage()
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

func inspectFile(root *string, key *string) {
	inKey := internalKey(*key)
	filename := filepath.Join(*root, inKey[0:3], inKey[3:6], inKey[6:])
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
}

func inspectDb(root *string, key *string, cmNum int) {
	inKey := internalKey(*key)
	dbIndex, err := getCmIndexFromKey(inKey, cmNum)
	if err != nil {
		log.Println(err)
	}
	dbPath := filepath.Join(*root, strconv.Itoa(dbIndex))
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

func inspect(root *string, key *string, sizeThreshold int, cmNum int) {
	cmRoot := filepath.Join(*root, contractManager)
	cmDb, err := leveldb.OpenFile(cmRoot, nil)
	if err != nil {
		fmt.Println("Open contract leveldb error: ", err)
		return
	}
	defer cmDb.Close()

	cdata, err := cmDb.Get([]byte(*key), nil)
	if err != nil {
		fmt.Println("Not find contract of ", *key, err)
		return
	}

	contract := Contract{}
	if err := json.Unmarshal(cdata, &contract); err != nil {
		fmt.Println("Decode contract error: ", err)
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
	fileStoreRoot := filepath.Join(*root, fileStoreName)
	if contract.Size > sizeThreshold {
		rsRoot := filepath.Join(fileStoreRoot, rsName)
		inspectFile(&rsRoot, key)
	} else {
		lsRoot := filepath.Join(fileStoreRoot, lsName)
		inspectDb(&lsRoot, key, cmNum)
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
	inspectDbNum := inspectCmd.Int("n", 2, "Chunk DB number")

	fileCmd := flag.NewFlagSet("file", flag.ExitOnError)
	fileKey := fileCmd.String("k", "", "block key")
	fileRoot := fileCmd.String("r", "", "Data root dir")

	dbCmd := flag.NewFlagSet("db", flag.ExitOnError)
	dbKey := dbCmd.String("k", "", "block key")
	dbRoot := dbCmd.String("r", "", "Data root dir")
	dbNum := dbCmd.Int("n", 2, "Chunk DB number")

	if len(os.Args) < 2 {
		printUsage(inspectCmd, fileCmd, dbCmd)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "inspect":
		inspectCmd.Parse(os.Args[2:])
	case "file":
		fileCmd.Parse(os.Args[2:])
	case "db":
		dbCmd.Parse(os.Args[2:])
	default:
		flag.Parse()
	}

	if inspectCmd.Parsed() {
		if *inspectKey == "" {
			fmt.Println("Please supply the key with -k option.")
			os.Exit(2)
		} else if *inspectRoot == "" {
			fmt.Println("Please supply the root dir with -r option.")
			os.Exit(3)
		}
		inspect(inspectRoot, inspectKey, *inspectSize, *inspectDbNum)
	} else if fileCmd.Parsed() {
		if *fileKey == "" {
			fmt.Println("Please supply the key with -k option.")
			os.Exit(2)
		} else if *fileRoot == "" {
			fmt.Println("Please supply the root dir with -r option.")
			os.Exit(3)
		}
		inspectFile(fileRoot, fileKey)
	} else if dbCmd.Parsed() {
		if *dbKey == "" {
			fmt.Println("Please supply the key with -k option.")
			os.Exit(2)
		} else if *dbRoot == "" {
			fmt.Println("Please supply the root dir with -r option.")
			os.Exit(3)
		}
		inspectDb(dbRoot, dbKey, *dbNum)
	} else { // if flag.Parsed()
		if true == *version {
			fmt.Printf("%s  %s\n", filepath.Base(os.Args[0]), VERSION)
		} else if true == *help {
			printUsage(inspectCmd, fileCmd, dbCmd)
		} else {
			fmt.Println("Unknow args ...")
			printUsage(inspectCmd, fileCmd, dbCmd)
		}
	}
}
