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
