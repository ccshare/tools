package main

import (
	"log"
)

// Token Generate Token
func Token(key string, op string) (string, error) {
	//
	log.Printf("Get token")
	retstr := "abcdefghijklmn"
	return retstr, nil
}

// Upload file
func Upload(key string) (string, error) {
	//
	log.Printf("upload file")
	retstr := "Upload file result"
	return retstr, nil
}

// Download file
func Download(key string) (string, error) {
	//
	log.Printf("download file")
	retstr := "Download file result"
	return retstr, nil
}
