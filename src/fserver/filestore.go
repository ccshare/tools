package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

type FileStore struct {
	db *leveldb.DB
}

func NewFileStore(path string) *FileStore {
	if "" == path {
		return nil
	}
	ldb, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &FileStore{db: ldb}
}

// Exist a key
func (fStore *FileStore) Exist(key string) (bool, error) {
	//
	return true, nil
}

// WriteDb db
func (fStore *FileStore) WriteDb(key string, value []byte) error {
	//
	log.Printf("FileStore.WriteDb: %s", key)
	err := fStore.db.Put([]byte(key), value, nil)
	if err != nil {
		log.Println("db.Put error:", key, err)
		return err
	}
	return nil
}

// ReadDb file
func (fStore *FileStore) ReadDb(key string) ([]byte, error) {
	//
	log.Printf("FileStore.ReadDb: %s", key)
	data, err := fStore.db.Get([]byte(key), nil)
	if err != nil {
		log.Println("db.Get error:", key, err)
		return []byte(""), err
	}
	return data, nil
}

// DeleteDb file
func (fStore *FileStore) DeleteDb(key string) error {
	//
	log.Printf("FileStore.DeleteDb: %s", key)
	err := fStore.db.Delete([]byte(key), nil)
	if err != nil {
		log.Println("db.Delete error:", key, err)
		return err
	}
	return nil
}
