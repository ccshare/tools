package main

import (
	"errors"
	"math/rand"
	"time"

	"github.com/golang/glog"
)

// TOKEN_RANDOM_LEN length
const tokenLen int = 24
const tokenExpireTime = 6 * 1000 * 1000 * 1000

var tokenLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type serverToken struct {
	key       string
	operation string
	now       time.Time
}

// FileServer struct
type FileServer struct {
	fStore *FileStore
	tokens map[string]serverToken
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Elapsed  for perf count
func Elapsed(start time.Time, funcName string) {
	glog.V(1).Infof("%s took %f seconds", funcName, time.Since(start).Seconds())
}

// NewFileServer func
func NewFileServer(path string) *FileServer {
	if path == "" {
		return nil
	}
	fileStore := NewFileStore(path)

	return &FileServer{fStore: fileStore, tokens: map[string]serverToken{}}
}

// Token Generate Token
func (fServer *FileServer) Token(key string, op string) (string, error) {
	//
	defer Elapsed(time.Now(), "FileServer.Token")
	tokenID := make([]byte, tokenLen)
	for i := 0; i < tokenLen; i++ {
		tokenID[i] = tokenLetters[rand.Intn(52)]
	}
	tokenStr := string(tokenID)
	glog.Infof("Gen %s token: %s %s", op, key, tokenStr)
	if _, exists := fServer.tokens[tokenStr]; exists {
		glog.Infof("token key[%s] already exist", tokenStr)
	}
	fServer.tokens[tokenStr] = serverToken{operation: op, now: time.Now()}
	return tokenStr, nil
}

// validate token
func (fServer *FileServer) validateToken(key, op string) (string, bool) {
	valid := false
	var msg string
	if token, exists := fServer.tokens[key]; exists == true {
		if time.Now().Sub(token.now) > tokenExpireTime {
			msg = "expired token"
		} else if token.operation != op {
			msg = "invalid operation"
		} else {
			msg = "valid"
			valid = true
		}
		delete(fServer.tokens, key)
	}

	return msg, valid
}

// Upload file
func (fServer *FileServer) Upload(token, key string, value []byte) (string, error) {
	defer Elapsed(time.Now(), "FileServer.Upload")
	if msg, valid := fServer.validateToken(token, "put"); valid == false {
		glog.Infoln("invalid token: ", msg)
		return msg, errors.New("invalid token")
	}
	glog.Infof("FileServer.Upload: %s %s", token, key)
	err := fServer.fStore.WriteDb(key, value)
	if err != nil {
		glog.Infoln("WriteDb failed:", token, key, err)
		return "Write DB failed", err
	}
	return "Upload success", nil
}

// Download file
func (fServer *FileServer) Download(token, key string) ([]byte, error) {
	defer Elapsed(time.Now(), "FileServer.Download")
	if msg, valid := fServer.validateToken(token, "get"); valid == false {
		glog.Infoln("invalid token: ", msg)
		return []byte(msg), nil
	}

	glog.Infof("FileServer.Download: %s %s", token, key)
	data, err := fServer.fStore.ReadDb(key)
	if err != nil {
		glog.Infoln("ReadDb failed:", token, key, err)
		return []byte("Read DB error"), err
	}

	return data, nil
}
