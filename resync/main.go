package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	appKey    = "GW_APP"
	bucketKey = "GW_BUCKET"
	userKey   = "GW_USER"
	storKey   = "GW_STOR"
	cacheKey  = "GW_CACHE_SETTING"
)

var (
	srcArg    string
	dstArg    string
	srcClient redisClient
	dstClient redisClient
	dbKey     = []string{
		"GW_APP",
		"GW_USER",
		"GW_STOR",
		"GW_BUCKET",
		"GW_CACHE_SETTING",
	}
	keyPrefix = []string{
		"SH::",
		"SH::",
	}
)

type redisClient struct {
	client  *redis.Client
	cluster *redis.ClusterClient
}

func (c *redisClient) Ping() (err error) {
	if c.client != nil {
		err = c.client.Ping().Err()
	} else if c.cluster != nil {
		err = c.cluster.Ping().Err()
	}
	return
}

func (c *redisClient) Keys(pattern string) (data []string, err error) {
	if c.client != nil {
		data, err = c.client.Keys(pattern).Result()
	} else if c.cluster != nil {
		data, err = c.cluster.Keys(pattern).Result()
	}
	return
}

func (c *redisClient) Dump(k string) (data []byte, err error) {
	if c.client != nil {
		data, err = c.client.Dump(k).Bytes()
	} else if c.cluster != nil {
		data, err = c.cluster.Dump(k).Bytes()
	}
	return
}

func (c *redisClient) TTL(k string) (data time.Duration, err error) {
	if c.client != nil {
		data, err = c.client.TTL(k).Result()
	} else if c.cluster != nil {
		data, err = c.cluster.TTL(k).Result()
	}
	return
}

func (c *redisClient) Restore(k, v string, ttl time.Duration) (err error) {
	if c.client != nil {
		err = c.client.Restore(k, ttl, v).Err()
	} else if c.cluster != nil {
		err = c.cluster.Restore(k, ttl, v).Err()
	}
	return
}

// parseRedisAddr parse redis passwd,ip,port,db from addr
func parseRedisAddr(addr string) (host, passwd string, db int, err error) {
	var u *url.URL
	u, err = url.ParseRequestURI(addr)
	if err != nil {
		return
	}

	if u.User != nil {
		var exists bool
		passwd, exists = u.User.Password()
		if !exists {
			passwd = u.User.Username()
		}
	}

	host = u.Host
	parts := strings.Split(u.Path, "/")
	if len(parts) == 1 {
		db = 0 //default redis db
	} else {
		db, err = strconv.Atoi(parts[1])
		if err != nil {
			db, err = 0, nil //ignore err here
		}
	}

	return
}

func initRedisClient(client *redisClient, addr string) (err error) {
	host, passwd, db, err := parseRedisAddr(addr)
	if err != nil { // invalid redis address
		return err
	}

	if strings.Contains(host, ",") { // redis cluster
		addrs := strings.Split(host, ",")
		client.cluster = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: passwd,
		})

	} else {
		client.client = redis.NewClient(&redis.Options{
			Addr:     host,
			Password: passwd,
			DB:       db,
		})

	}
	return client.Ping()
}

func keySync(key []string) error {
	for _, k := range key {
		data, err := srcClient.Dump(k)
		if err != nil {
			return fmt.Errorf("dump %s error %w", k, err)
		}
		err = dstClient.Restore(k, string(data), 0)
		if err != nil {
			return fmt.Errorf("dump %s error %w", k, err)
		}
	}
	return nil
}

func prefixSync(prefix []string) error {
	for _, p := range prefix {
		keys, err := srcClient.Keys(p)
		if err != nil {
			return fmt.Errorf("keys %s error %w", p, err)
		}
		for _, k := range keys {
			data, err := srcClient.Dump(k)
			if err != nil {
				return fmt.Errorf("dump %s error %w", k, err)
			}
			err = dstClient.Restore(k, string(data), 0)
			if err != nil {
				return fmt.Errorf("dump %s error %w", k, err)
			}
		}
	}
	return nil
}

func main() {
	flag.StringVar(&srcArg, "src", "redis://127.0.0.1:6379/1", "src redis address")
	flag.StringVar(&dstArg, "dst", "redis://127.0.0.1:6379/2", "dst redis address")
	flag.Parse()

	err := initRedisClient(&srcClient, srcArg)
	if err != nil {
		log.Fatal("init src redis client", err)
	}
	err = initRedisClient(&dstClient, dstArg)
	if err != nil {
		log.Fatal("init dst redis client", err)
	}

	err = keySync(dbKey)
	if err != nil {
		log.Fatal("key sync error", err)
	}

	err = prefixSync(keyPrefix)
	if err != nil {
		log.Fatal("prefix sync error", err)
	}

}
