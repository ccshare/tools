package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
)

// parseRedisAddr ...
func parseRedisAddr(addr string) (host, password string, db int, err error) {
	if len(addr) < 10 { // invalid redis address
		err = fmt.Errorf("invalid redis addr %s", addr)
		return
	}
	// redis:  pwd@host/db
	if !strings.HasPrefix(addr, "redis://") {
		addr = "redis://" + addr
	}
	var u *url.URL
	u, err = url.Parse(addr)
	if err != nil {
		return
	}

	if u.User != nil {
		var exists bool
		password, exists = u.User.Password()
		if !exists {
			password = u.User.Username()
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

func dbInit(addr string) (*redis.Client, error) {
	host, passwd, db, err := parseRedisAddr(addr)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(&redis.Options{
		Addr:     host,
		Password: passwd,
		DB:       db,
	}), nil
}
