package main

import (
	"testing"
)

func Test_internalKey(t *testing.T) {
	// sha1sum of key abcdefg is 2fb5e13419fc89246865e7a324f476ec624e8740
	inKey := internalKey("abcdefg")
	if inKey == "2fb5e13419fc89246865e7a324f476ec624e8740" {
		t.Log("success")
	} else {
		t.Error("failed", inKey)
	}

}

func Test_getChunkKeyByIndex_0(t *testing.T) {
	chunkKey := getChunkKeyByIndex("abcdefg", 0)
	if chunkKey == "abcdefg-000000" {
		t.Log("success")
	} else {
		t.Error("failed", chunkKey)
	}

}

func Test_getChunkKeyByIndex_1(t *testing.T) {
	chunkKey := getChunkKeyByIndex("abcdefg", 1)
	if chunkKey == "abcdefg-000001" {
		t.Log("success")
	} else {
		t.Error("failed", chunkKey)
	}

}

func Test_getChunkKeyByIndex_2(t *testing.T) {
	chunkKey := getChunkKeyByIndex("abcdefg", 2)
	if chunkKey == "abcdefg-000002" {
		t.Log("success")
	} else {
		t.Error("failed", chunkKey)
	}

}

func Test_getCmIndexFromKey_0(t *testing.T) {
	// Should fail if arg 1 is not a hex string
	if _, err := getCmIndexFromKey("cgd1f1d9578870d34b168b64e0b8465fca36d533", 2); err == nil {
		t.Error("failed")
	} else {
		t.Log("success")
	}
}

func Test_getCmIndexFromKey_1(t *testing.T) {
	index, err := getCmIndexFromKey("cfd1f1d9578870d34b168b64e0b8465fca36d533", 2)
	if err != nil {
		t.Error("failed", err)
	} else if index != 0 {
		t.Log("success")
	} else {
		t.Error("failed", index)
	}
}
