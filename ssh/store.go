package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
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

func dbInit(addr, markerKey, markerField string) (*redis.Client, string, error) {
	host, passwd, db, err := parseRedisAddr(addr)
	if err != nil {
		return nil, "", err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: passwd,
		DB:       db,
	})

	marker, err := client.HGet(markerKey, markerField).Result()
	if err != nil && err != redis.Nil {
		return nil, "", err
	}
	return client, marker, nil
}

func findMarkerFromFile(name string) string {
	fd, err := os.Open(name)
	if err != nil {
		return ""
	}
	defer fd.Close()

	line := ""
	var cursor int64 = 0
	stat, _ := fd.Stat()
	filesize := stat.Size()
	for {
		cursor--
		fd.Seek(cursor, io.SeekEnd)

		char := make([]byte, 1)
		fd.Read(char)

		if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

		if cursor == -filesize { // stop if we are at the begining
			break
		}
	}
	if line == "" {
		return ""
	}

	fields := strings.Split(line, " ")
	if len(fields) > 12 && fields[1] != "date" {
		return fmt.Sprintf("%s %s", fields[0], fields[1])
	}
	return ""
}
