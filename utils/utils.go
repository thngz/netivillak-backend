package utils

import (
	"math/rand"
	"strings"
)

func CreateRandomId(length int) string {
	var sb strings.Builder
	a := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	for range length {
		r := rand.Intn(len(a))
		sb.WriteByte(a[r])
	}
    
	return sb.String()
}
