package x

import (
	"math/rand"
	"time"
)

const dict = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/+"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func CustomRandomString(size int, words string) string {
	ret := make([]byte, size)
	l := len(words)
	for i := 0; i < size; i++ {
		d := rand.Int() % l
		ret[i] = words[d]
	}
	return string(ret)
}

func RandomString(size int) string {
	return CustomRandomString(size, dict)
}
