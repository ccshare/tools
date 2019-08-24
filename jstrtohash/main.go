package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/go-redis/redis"
)

func convert(s, p, j, h string, db int) error {
	client := redis.NewClient(&redis.Options{
		Addr:       s,
		Password:   p,
		DB:         db,
		MaxRetries: 2,
	})

	if _, err := client.Ping().Result(); err != nil {
		return err
	}

	jstr := client.Get(j)
	//if err != nil {
	//	return fmt.Errorf("get %s error %s\n", j, err)
	//}
	fmt.Println(jstr)

	apps := map[string]interface{}{}
	if err := json.Unmarshal([]byte(jstr.Val()), &apps); err != nil {
		return err
	}
	for k, v := range apps {
		if jsv, err := json.Marshal(v); err != nil {
			fmt.Printf("marshal v to json error: %s\n", err)
			continue
		} else {
			v := client.HSet(h, k, jsv)
			fmt.Printf("HSET %s %v\n", k, v.Val())
		}
	}

	return nil
}

func main() {
	server := flag.String("addr", "192.168.55.2:6379", "redis server address")
	passwd := flag.String("passwd", "", "redis passwd")
	db := flag.Int("db", 0, "redis db")
	jsonkey := flag.String("str", "GW_APPIDS", "json str key")
	hashkey := flag.String("hash", "GW_APP", "hash key")
	flag.Parse()
	fmt.Printf("convert %s json str %s to hash %s\n", *server, *jsonkey, *hashkey)
	if err := convert(*server, *passwd, *jsonkey, *hashkey, *db); err != nil {
		fmt.Printf("convert error: %s\n", err)
	}
}
