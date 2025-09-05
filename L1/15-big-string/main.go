package main

import (
	"math/rand"
	"strings"
)

var justString string

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func createHugeString(size int) string {
	var builder strings.Builder
	builder.Grow(size)
	for i := 0; i < size; i++ {
		builder.WriteRune(letters[rand.Intn(len(letters))])
	}
	return builder.String()
}

func someFunc() {
	v := createHugeString(1 << 8)
	// justString = v[:100]
	// При создании слайса все еще будет ссылка на большую строку
	// из-за чего GC не сможет её удалить из памяти
	// Поэтому необходимо создать новую строку с той частью, что нам нужна (скопировать по сути)
	justString = string(v[:100])
}

func main() {
	someFunc()
}
