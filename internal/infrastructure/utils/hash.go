package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"io"
)

// Hash 用于计算问题的hash值
func Hash(question string) (hashVal string, err error) {
	hasher := sha256.New()
	_, err = io.WriteString(hasher, question)
	if err != nil {
		err = fmt.Errorf("io.WriteString(hasher, question) err: %v", zap.Error(err))
		return
	}
	// 使用 hex.EncodeToString 将二进制哈希结果转换为十六进制字符串
	hashVal = hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("\n问题: %s\n", question)
	fmt.Printf("Hash值: %s\n", hashVal)
	return
}
