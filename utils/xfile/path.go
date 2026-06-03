package xfile

import (
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizePath 规范化路径
func NormalizePath(path string) string {
	// 清理路径中的反斜杠
	path = filepath.Clean(path)
	// 处理 Unix 风格的 Windows 路径：/C:/path
	if runtime.GOOS == "windows" {
		// 方式1: 使用 filepath.VolumeName
		if strings.HasPrefix(path, "/") && len(path) > 2 && path[2] == ':' {
			path = path[1:] // 移除前导斜杠
		}
	}
	return path
}
