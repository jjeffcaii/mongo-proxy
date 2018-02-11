package x

import (
	"math/rand"
	"time"
)

const randChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/+"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func CustomRandomString(size int, words string) string {
	ret := make([]byte, size)
	for i := 0; i < size; i++ {
		foo := rand.Intn(len(words))
		ret[i] = words[foo]
	}
	return string(ret)
}

func RandomString(size int) string {
	return CustomRandomString(size, randChars)
}
