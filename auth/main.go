package main

import (
	"fmt"

	"github.com/casbin/casbin"
)

func main() {
	e, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		panic(err.Error())
	}

	sub := "alice"
	obj := "data1"
	act := "read"

	if res := e.Enforce(sub, obj, act); res {
		fmt.Println("Ok")
	} else {
		fmt.Println("Failed")
	}
}
