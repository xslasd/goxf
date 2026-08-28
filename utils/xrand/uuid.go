package xrand

import (
	"strings"

	"github.com/google/uuid"
)

// UUIDv7 生成 RFC 9562 标准的 Version 7 UUID（基于毫秒级时间戳有序递增）
func UUIDv7() string {
	u, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return u.String()
}

// UUIDv7Simple 生成去除连字符的 32 位小写 UUIDv7 字符串
func UUIDv7Simple() string {
	return strings.ReplaceAll(UUIDv7(), "-", "")
}

// UUID 生成标准随机 UUID (v4)
func UUID() string {
	return uuid.New().String()
}
