package main

import (
	"github.com/golang/glog"
	"github.com/syndtr/goleveldb/leveldb"
)

//FileStore Presenting a FileStore type
type FileStore struct {
	db *leveldb.DB
}

// NewFileStore constructor
func NewFileStore(path string) *FileStore {
	if "" == path {
		return nil
	}
	ldb, err := leveldb.OpenFile(path, nil)
	if err != nil {
		glog.Infoln(err)
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
	glog.Infof("FileStore.WriteDb: %s", key)
	err := fStore.db.Put([]byte(key), value, nil)
	if err != nil {
		glog.Infoln("db.Put error:", key, err)
		return err
	}
	return nil
}

// ReadDb file
func (fStore *FileStore) ReadDb(key string) ([]byte, error) {
	//
	glog.Infof("FileStore.ReadDb: %s", key)
	data, err := fStore.db.Get([]byte(key), nil)
	if err != nil {
		glog.Infoln("db.Get error:", key, err)
		return []byte(""), err
	}
	return data, nil
}

// DeleteDb file
func (fStore *FileStore) DeleteDb(key string) error {
	//
	glog.Infof("FileStore.DeleteDb: %s", key)
	err := fStore.db.Delete([]byte(key), nil)
	if err != nil {
		glog.Infoln("db.Delete error:", key, err)
		return err
	}
	return nil
}
