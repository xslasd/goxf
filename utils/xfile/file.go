package xfile

import (
	"os"
	"path/filepath"
)

// Exists 检查文件或目录是否存在
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err // 其他错误（如权限问题）
}

// LookupFilesByDirs 递归查找目录下的所有文件
// root: 查找的根目录
// exts: 只筛选指定文件扩展名
func LookupFilesByDirs(root string, exts []string) ([]string, error) {
	files := make([]string, 0)
	f, err := os.Open(root)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	extsLen := len(exts)
	for _, name := range names {
		filePath := filepath.Join(root, name)
		isDir, _ := IsDirectory(filePath)
		if isDir {

			list, err := LookupFilesByDirs(filePath, exts)
			if err != nil {
				return files, err
			}
			files = append(files, list...)
			continue
		}
		if extsLen > 0 {
			ext := filepath.Ext(name)
			for _, e := range exts {
				if e == ext {
					files = append(files, filePath)
				}
			}
		} else {
			files = append(files, filePath)
		}
	}
	return files, nil
}
