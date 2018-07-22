package main

import (
	"testing"
)

func Test_generate(t *testing.T) {
	if _, err := generate(); err != nil {
		t.Log("success")
	}
}
