package main

import (
	"testing"
)

func TestFailedExist(t *testing.T) {
	filestore := FileStore{}
	// func (fStore *FileStore) Exist(key string) (bool, error)
	if exist, _ := filestore.Exist(""); exist == false {
		t.Log("success")
	} else {
		t.Error("failed")
	}

}

func TestSuccessExist(t *testing.T) {
	filestore := FileStore{}
	// func (fStore *FileStore) Exist(key string) (bool, error)
	if exist, _ := filestore.Exist("key"); exist == true {
		t.Log("success")
	} else {
		t.Error("failed")
	}

}
