package main

import (
	"flag"
	"fmt"

	"github.com/casbin/casbin/v2"
	mongodbadapter "github.com/casbin/mongodb-adapter/v2"
	redisadapter "github.com/casbin/redis-adapter/v2"
)

var version = "unknown"

func casbinFile(s, o, a string) (bool, error) {
	e, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		return false, err
	}

	return e.Enforce(s, o, a)
}

func casbinRedis(s, o, a string) (bool, error) {
	// Initialize a Redis adapter and use it in a Casbin enforcer:
	adp := redisadapter.NewAdapter("tcp", "192.168.55.2:6379") // Your Redis network and address.
	// Use the following if Redis has password like "123"
	//a := redisadapter.NewAdapterWithPassword("tcp", "127.0.0.1:6379", "123")
	e, err := casbin.NewEnforcer("model.conf", adp)
	if err != nil {
		return false, err
	}
	// Load the policy from DB.
	e.LoadPolicy()

	// Modify the policy.
	e.AddPolicy(s, o, a)
	// e.RemovePolicy(s, o, a)

	// Save the policy back to DB.
	e.SavePolicy()

	// Check the permission.
	return e.Enforce(s, o, a)
}

func casbinMongo(s, o, a string) (bool, error) {
	// Initialize a MongoDB adapter and use it in a Casbin enforcer:
	// The adapter will use the database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	adp := mongodbadapter.NewAdapter("mongodb://casbin_user:password@192.168.55.2:27017")

	// Or you can use an existing DB "abc" like this:
	// The adapter will use the table named "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	// a := mongodbadapter.NewAdapter("mongodb://user:pass@127.0.0.1:27017/abc")

	e, err := casbin.NewEnforcer("model.conf", adp)
	if err != nil {
		return false, err
	}

	// Load the policy from DB.
	e.LoadPolicy()

	// Modify the policy.
	// e.AddPolicy(s, o, a)
	// e.RemovePolicy(s, o, a)
	//e.SavePolicy()

	// Check the permission.
	return e.Enforce(s, o, a)
}

func main() {
	sub := flag.String("sub", "alice", "sub")
	obj := flag.String("obj", "data1", "obj")
	act := flag.String("act", "read", "act")
	flag.Parse()

	//res, err := casbinFile(*sub, *obj, *act)
	//res, err := casbinRedis(*sub, *obj, *act)
	res, err := casbinMongo(*sub, *obj, *act)
	if err != nil {
		fmt.Println("error: ", err)
	} else {
		fmt.Printf("%v, %v, %v, %v\n", *sub, *obj, *act, res)
	}

}
