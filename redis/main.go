package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func main() {
	server := flag.String("addr", "192.168.55.2:6379", "redis server address")
	passwd := flag.String("passwd", "", "redis passwd")
	db := flag.Int("db", 0, "redis db")
	num := flag.Int("n", 10, "hash field num")
	size := flag.Int("z", 32, "hash value size")
	key := flag.String("key", "hash_test_key", "hash key")
	flag.Parse()

	fmt.Printf("write %d fields to %s with size %d\n", *num, *key, *size)
	client := redis.NewClient(&redis.Options{
		Addr:       *server,
		Password:   *passwd,
		DB:         *db,
		MaxRetries: 2,
	})
	defer client.Close()

	client.Del(*key)
	data := make([]byte, *size)
	if n, err := rand.Read(data); err != nil {
		fmt.Println("gen rand data error: ", err)
		return
	} else if n < *size {
		fmt.Printf("gen rand data size invalid %d < %d\n ", n, *size)
		return
	}

	strData := string(data)
	start := time.Now()
	for i := 0; i < *num; i++ {
		f := fmt.Sprintf("key-index-%04d", i)
		client.HSet(*key, f, strData)
	}
	fmt.Printf("hset    elapsed %d\n", time.Since(start).Milliseconds())

	start = time.Now()
	if err := client.HGetAll(*key).Err(); err != nil {
		fmt.Println("hgetall error: ", err)
	}
	fmt.Printf("hgetall elapsed %d\n", time.Since(start).Milliseconds())
}
