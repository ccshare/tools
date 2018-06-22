package main

import (
	"testing"
)

func TestFailedToken(t *testing.T) {
	// func token(serverURL string, entryKey string, entryOp string) (string, error)
	if _, err := token("http://127.0.0.1", "key", "put"); err != nil {
		t.Log("success")
	}

}

func TestSuccessToken(t *testing.T) {
	// func token(serverURL string, entryKey string, entryOp string) (string, error)
	if _, err := token("http://127.0.0.1", "key", "put"); err == nil {
		t.Log("success")
	} else {
		t.Error("failed")
	}

}
