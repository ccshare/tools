package main

import (
	"fmt"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

func validateToken(tokenString string) bool {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// signSecretKey is a []byte containing your secret, e.g. []byte("my_secret_key")
		return signSecretKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["sub"], claims["exp"])
	} else {
		fmt.Println(err)
	}

	return token.Valid
}

func Test_createToken(t *testing.T) {
	token, err := createToken("username")
	if err != nil {
		t.Errorf("createToken %s", err)
	}

	if v := validateToken(token); v != true {
		t.Error("validate token failed")
	}

}
