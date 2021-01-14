package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"math"
	"strings"

	"github.com/rryqszq4/go-murmurhash"
	"github.com/tenfyzhong/cityhash"
)

var (
	funType uint
	content string
	// CharacterSet consists of 62 characters [0-9][A-Z][a-z].
	CharacterSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

const (
	_base = 62
)

// base62Encode32 returns a base62 representation as string of the given integer number.
func base62Encode32(num uint32) string {
	b := make([]byte, 0)

	// loop as long the num is bigger than zero
	for num > 0 {
		// receive the rest
		r := math.Mod(float64(num), float64(_base))

		// devide by Base
		num /= _base

		// append chars
		b = append([]byte{CharacterSet[int(r)]}, b...)
	}

	return string(b)
}

// base62Encode64 returns a base62 representation as string of the given integer number.
func base62Encode64(num uint64) string {
	b := make([]byte, 0)

	// loop as long the num is bigger than zero
	for num > 0 {
		// receive the rest
		r := math.Mod(float64(num), float64(_base))

		// devide by Base
		num /= _base

		// append chars
		b = append([]byte{CharacterSet[int(r)]}, b...)
	}

	return string(b)
}

// base62Decode returns a integer number of a base62 encoded string.
func base62Decode(s string) (int, error) {
	var r, pow int

	// loop through the input
	for i, v := range s {
		// convert position to power
		pow = len(s) - (i + 1)

		// IndexRune returns -1 if v is not part of CharacterSet.
		pos := strings.IndexRune(CharacterSet, v)

		if pos == -1 {
			return pos, errors.New("invalid character: " + string(v))
		}

		// calculate
		r += pos * int(math.Pow(float64(_base), float64(pow)))
	}

	return int(r), nil
}

func md5Encode(s string) []byte {
	hash := md5.New()
	hash.Write([]byte(s))
	return hash.Sum(nil)
}

func mm3Encode(s string) uint32 {
	return murmurhash.MurmurHash3_x86_32([]byte("key"), 8)
}

func main() {
	flag.UintVar(&funType, "t", 1, "hash type")
	flag.StringVar(&content, "s", "test-string", "string content")
	flag.Parse()

	md5Str := md5Encode(content)

	fmt.Printf("md5hash: hex   : %v\n", hex.EncodeToString(md5Str))
	fmt.Printf("md5hash: base64: %v\n", base64.URLEncoding.EncodeToString(md5Str))

	mm3x86_128 := murmurhash.MurmurHash3_x86_128([]byte(content), 8)
	fmt.Printf("mm3hash: %v: %v\n", mm3x86_128, base62Encode32(mm3x86_128[0]))

	mm3x64_128 := murmurhash.MurmurHash3_x64_128([]byte(content), 8)
	fmt.Printf("mm3hash: %v: %v\n", mm3x64_128, base62Encode64(mm3x64_128[0]))

	mm1V := murmurhash.MurmurHash1([]byte(content), 8)
	fmt.Printf("mm1hash: %T, %v: %v\n", mm1V, mm1V, base62Encode32(mm1V))

	mm2V := murmurhash.MurmurHash2([]byte(content), 8)
	fmt.Printf("mm2hash: %T, %v: %v\n", mm2V, mm2V, base62Encode32(mm2V))

	mm3x86_32 := murmurhash.MurmurHash3_x86_32([]byte(content), 8)
	fmt.Printf("mm3hash: %T, %v: %v\n", mm3x86_32, mm3x86_32, base62Encode32(mm3x86_32))

	city32 := cityhash.CityHash32([]byte(content))
	fmt.Printf("city32 : %T, %v: %v\n", city32, city32, base62Encode32(city32))

	city64 := cityhash.CityHash64([]byte(content))
	fmt.Printf("city64 : %T, %v: %v\n", city64, city64, base62Encode64(city64))

	city128 := cityhash.CityHash128([]byte(content))
	fmt.Printf("city128: %T, %v: %v\n", city128, city128, base62Encode64(city64))
}

/*
661f8009fa8e56a9d0e94a0a644397d7
Zh-ACfqOVqnQ6UoKZEOX1w==
*/
