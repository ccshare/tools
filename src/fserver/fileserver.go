package main

import (
	"log"
)

type FileServer struct {
	fStore *FileStore
}

func NewFileServer(path string) *FileServer {
	if path == "" {
		return nil
	}
	fileStore := NewFileStore(path)

	return &FileServer{fStore: fileStore}
}

// Token Generate Token
func (fServer *FileServer) Token(key string, op string) (string, error) {
	//
	log.Printf("Get token")
	retstr := "abcdefghijklmn"
	return retstr, nil
}

// Upload file
func (fServer *FileServer) Upload(key string, value []byte) (string, error) {
	//
	log.Printf("FileServer.Upload")
	err := fServer.fStore.WriteDb(key, value)
	if err != nil {
		log.Println("FileServer.Upload", err)
		return "Write DB failed", err
	}
	return "Upload success", nil
}

// Download file
func (fServer *FileServer) Download(key string) ([]byte, error) {
	//
	log.Printf("FileServer.Download")
	data, err := fServer.fStore.ReadDb(key)
	if err != nil {
		log.Println("ReadDb failed", err)
		return []byte("Read DB error"), err
	}

	return data, nil
}
