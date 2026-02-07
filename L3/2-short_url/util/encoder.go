package util

import (
	"errors"
	"strings"
)

const base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func EncodeBase58(num int64) string {
	base := int64(58)
	// 11 is a max length could be for in64 and base 58
	maxLength := 11
	result := [11]rune{}
	length := 0
	if num == 0 {
		return string(base58Chars[0])
	}
	for num > 0 {
		remainder := num % base
		result[maxLength-length-1] = rune(base58Chars[remainder])
		num = num / base
		length++
	}
	return string(result[maxLength-length:])
}

func DecodeBase58(text string) (int64, error) {
	var result int64
	base := int64(58)
	for _, char := range text {
		index := int64(strings.IndexRune(base58Chars, char))
		if index == -1 {
			return -1, errors.New("invalid char to decode in base 58")
		}
		result = result*base + index
	}
	return result, nil
}
