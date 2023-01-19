package models

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

//得到文件，如果没有创建一个新文件
//flag false表示存在  true表示新建
func GetFileOrCreate(fileName string) (file1 *os.File, flagBool bool, err error) {
	filePath := "./" + fileName
	exist, _ := PathExists(filePath)
	flag := false
	if exist {
		flag = false
	} else {
		//新建
		flag = true
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	return file, flag, err
}

//写文件
func WriteFile(fileName, content string) {
	file, flag, err := GetFileOrCreate(fileName)
	if flag {
		fmt.Println("文件新建")
	} else {
		fmt.Println("文件已经存在")
	}
	if err != nil {
		fmt.Println("文件打开失败", err)
		return
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString(content)
	//Flush将缓存的文件真正写入到文件中
	write.Flush()
}

//写文件附加
func WriteFileAppend(fileName, content string) {
	exist, _ := PathExists(fileName)
	var err error
	var file *os.File
	if exist {
		fmt.Println("使用已经存在文件")
		file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0666)
	} else {
		fmt.Println("创建新文件")
		file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	}
	if err != nil {
		fmt.Println("文件打开失败", err)
		return
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString(content)
	//Flush将缓存的文件真正写入到文件中
	write.Flush()
}

//多线程写入字符串

//读取文件内容
//适合小文件读取
func GetSmallFileContent(fileName string) string {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return ""
	}
	//fmt.Println(string(content))
	return string(content)
}

//移除文件
func RemoveFile(fileName string) {
	exist, _ := PathExists(fileName)
	if exist {
		os.Remove(fileName)
	}
}
