/*
@ Author:       Wang XiaoQiang
@ Github:       https://github.com/wangxiaoqiange
@ File:         tokenizer.go
@ Create Time:  2019-08-02 11:08
@ Software:     GoLand
*/

package tools

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GetMD5Hash(text string, refresh bool) string {
	hash := md5.New()
	if refresh {
		text += randomString(4)
	}
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}
