package utils

import (
	"strings"
)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func IdToShortURL(n uint) string {
	if n == 0 {
		return string(charset[0])
	}

	var shortURL strings.Builder
	for n > 0 {
		remainder := n % 62
		shortURL.WriteByte(charset[remainder])
		n /= 62
	}

	runes := []rune(shortURL.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func ShortURLToID(shortURL string) uint {
	var id uint
	for _, c := range shortURL {
		switch {
		case 'a' <= c && c <= 'z':
			id = id*62 + uint(c-'a')
		case 'A' <= c && c <= 'Z':
			id = id*62 + uint(c-'A'+26)
		case '0' <= c && c <= '9':
			id = id*62 + uint(c-'0'+52)
		}
	}
	return id
}
