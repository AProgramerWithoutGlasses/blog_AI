package code_infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
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
	// 使用 hex.EncodeToString 将二进制哈希结果转换为十六进制字符串
	hashVal = hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("SHA-256 Hash: %s\n", hashVal)
	return
}
