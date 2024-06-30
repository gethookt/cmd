package id

import (
	"math/rand"
	"strings"
)

const (
	all = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func Gen(n int) string {
	return gen(all, 6, 1<<6-1, 63/6, n)
}

// source: https://stackoverflow.com/questions/22892120/
func gen(chars string, idxBits uint, idxMask int64, idxMax int, n int) string {
	var sb strings.Builder

	sb.Grow(n)

	for i, cache, remain := n-1, rand.Int63(), idxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		if idx := int(cache & idxMask); idx < len(chars) {
			sb.WriteByte(chars[idx])
			i--
		}
		cache >>= idxBits
		remain--
	}

	return sb.String()
}
