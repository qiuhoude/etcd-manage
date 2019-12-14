package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// 获取执行路径
func GetRootDir() string {
	// 文件不存在获取执行路径
	//file, err := filepath.Abs(filepath.Dir(os.Args[0]))
	file, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		file = fmt.Sprintf(".%s", string(os.PathSeparator))
	} else {
		file = fmt.Sprintf("%s%s", file, string(os.PathSeparator))
	}
	//fmt.Println("file-->", file)
	return file
}

// PathExists 判断文件或目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	// 判断错误是不是 文件不存在错误
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
