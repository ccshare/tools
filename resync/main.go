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
	srcClient redisClient
	dstClient redisClient
	srcArg    string
	dstArg    string
	force     bool
	keyMode   bool
	dbKey     string
	gwDBKey   = []string{
		"GW_APP",
		"GW_USER",
		"GW_STOR",
		"GW_BUCKET",
		"GW_CACHE_SETTING",
	}
	dbKeyPattern = []string{
		"SH::*",
		"SZ::*",
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

func (c *redisClient) Del(k string) (err error) {
	if c.client != nil {
		err = c.client.Del(k).Err()
	} else if c.cluster != nil {
		err = c.cluster.Del(k).Err()
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
			if err == redis.Nil {
				log.Printf("key %s not exists", k)
				continue
			}
			return fmt.Errorf("dump %s error %w", k, err)
		}
		if force {
			dstClient.Del(k)
		}
		err = dstClient.Restore(k, string(data), 0)
		if err != nil {
			log.Printf("restore %s error %s", k, err)
		} else {
			fmt.Printf("success sync: %s\n", k)
		}
	}
	return nil
}

func patternSync(prefix []string) error {
	for _, p := range prefix {
		keys, err := srcClient.Keys(p)
		if err != nil {
			return fmt.Errorf("keys %s error %w", p, err)
		}
		if err := keySync(keys); err != nil {
			log.Printf("sync pattern error %s", err)
		}
	}
	return nil
}

func main() {
	flag.StringVar(&srcArg, "src", "redis://192.168.55.2:6379/8", "src redis address")
	flag.StringVar(&dstArg, "dst", "redis://127.0.0.1:6379/2", "dst redis address")
	flag.StringVar(&dbKey, "k", "RMR_ZONE_INFO", "redis key to sync")
	flag.BoolVar(&force, "force", false, "force sync, will overwrite exists key")
	flag.BoolVar(&keyMode, "K", false, "only sync -k redis-key")
	flag.Parse()

	err := initRedisClient(&srcClient, srcArg)
	if err != nil {
		log.Fatal("init src redis client error: ", err)
	}
	err = initRedisClient(&dstClient, dstArg)
	if err != nil {
		log.Fatal("init dst redis client error: ", err)
	}

	err = keySync([]string{dbKey})
	if err != nil {
		log.Printf("sync key %s error: %s", dbKey, err)
	}

	if keyMode {
		return
	}

	err = keySync(gwDBKey)
	if err != nil {
		log.Printf("sync gw keys error: %s", err)
	}

	err = patternSync(dbKeyPattern)
	if err != nil {
		log.Printf("sync pattern error:%s ", err)
	}

}
