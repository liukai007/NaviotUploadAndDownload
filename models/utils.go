package models

import (
	"os"
	"runtime"
)

//是否是linux系统
func IsLinux() bool {
	sysType := runtime.GOOS
	if sysType == "linux" {
		// LINUX系统
		return true
	}
	return false
}

//PathExists 判断一个文件或文件夹是否存在
//输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
