package code_infrastructure

import (
	"crypto/sha256"
	"fmt"
	"io"
)

func Hash(key string) (hashVal string, err error) {
	hasher := sha256.New()
	_, err = io.WriteString(hasher, key)
	if err != nil {
		fmt.Println("hash.hash() io.WriteString() err:", err)
		return
	}
	hashVal = string(hasher.Sum(nil))
	fmt.Printf("SHA-256 Hash: %s\n", hashVal)
	return
}
