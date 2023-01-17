package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math"
	"os"
)

const SliceBytes = 1024 * 1024 * 1 // 分片大小1MB

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给文件是否存在
func IsFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

// GetFileSize 获取文件大小 单位是字节
func GetFileSize(path string) int64 {
	fh, err := os.Stat(path)
	if err != nil {
		fmt.Printf("读取文件%s失败, err: %s\n", path, err)
	}
	return fh.Size()
}

//MD5计算
func FileMD5(filePath string) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	defer file.Close()
	return hex.EncodeToString(hash.Sum(nil)), nil
}

//随机生成，得到UUID
func GetUUID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		fmt.Println("生成UUID失败")
		return "", err
	}
	return uuid.String(), err
}

//计算文件切片数量
func GetSliceNum(filesize int64) int64 {
	// 计算文件切片数量
	sliceNum := int(math.Ceil(float64(filesize) / float64(SliceBytes)))
	return int64(sliceNum)
}

//文件目录不存在就创建
func DirCreate(dirPathStr string) {
	if IsFile(dirPathStr) {
		fmt.Println("这个路径是一个文件直接路径")
	}
	if !IsDir(dirPathStr) {
		err := os.Mkdir(dirPathStr, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
}

//删除文件
func DelFile(filePathStr string) {
	if !IsFile(filePathStr) {
		fmt.Println("文件已经不存在")
	}
	err := os.Remove(filePathStr)
	if err != nil {
		//如果删除失败则输出 file remove Error!
		fmt.Println("文件删除失败可能文件正在被使用")
		//输出错误详细信息
		fmt.Printf("%s", err)
	} else {
		//如果删除成功则输出 file remove OK!
		fmt.Println("文件删除成功")
	}
}

//删除文件夹和里面的内容
func DelFileDir(dirPathStr string) {
	err := os.RemoveAll(dirPathStr)
	if err != nil {
		fmt.Println(err)
		fmt.Println("文件夹删除失败")
	} else {
		fmt.Println("文件夹删除成功")
	}
}
