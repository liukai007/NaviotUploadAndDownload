package models

import (
	"bufio"
	"fmt"
	"github.com/axgle/mahonia"
	"io"
	"os"
	"strings"
)

//读取txt文件，自动判断编码格式
func GetStrList(path string) ([]string, error) {
	f, err := os.Open(path) //打开目录路径txt文件
	if err != nil {
		return nil, err
	}
	defer f.Close() //最后关闭文件

	r := bufio.NewReader(f)

	buf := make([]byte, 1024)
	var res string

	_, err = r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	res = GetStrCoding(buf)

	if res == "UTF8" {
		return GetTextContentUTF8(path), nil
	} else {
		return GetTextContentGbk(path), nil
	}
}

//读取gbk编码格式的文件
func GetTextContentGbk(txtPath string) []string {
	f, err := os.Open(txtPath) //打开目录路径txt文件
	if err != nil {
		fmt.Println("err=", err)
	}
	defer f.Close()

	decoder := mahonia.NewDecoder("gbk")
	r := bufio.NewReader(decoder.NewReader(f))

	chunks := []byte{}
	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == n {
			break
		}
		chunks = append(chunks, buf[:n]...)
	}

	nameStr := strings.ReplaceAll(string(chunks), "\r\n", ",")
	return strings.Split(nameStr, ",")
}

//读取utf8编码格式的文件
func GetTextContentUTF8(txtPath string) []string {
	f, err := os.Open(txtPath) //打开目录路径txt文件
	if err != nil {
		fmt.Println("err=", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	chunks := []byte{}
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == n {
			break
		}
		chunks = append(chunks, buf[:n]...)
	}

	nameStr := strings.ReplaceAll(string(chunks), "\r\n", ",")
	return strings.Split(nameStr, ",")
}

const (
	GBK     string = "GBK"
	UTF8    string = "UTF8"
	UNKNOWN string = "UNKNOWN"
)

//判断文本文件格式
func GetStrCoding(data []byte) string {
	if isUtf8(data) == true {
		return UTF8
	} else if isGBK(data) == true {
		return GBK
	} else {
		return UNKNOWN
	}
}

func isGBK(data []byte) bool {
	length := len(data)
	var i = 0
	for i < length {
		if data[i] <= 0x7f {
			//编码0~127,只有一个字节的编码，兼容ASCII码
			i++
			continue
		} else {
			//大于127的使用双字节编码，落在gbk编码范围内的字符
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func preNUm(data byte) int {
	var mask byte = 0x80
	var num int = 0
	//8bit中首个0bit前有多少个1bits
	for i := 0; i < 8; i++ {
		if (data & mask) == mask {
			num++
			mask = mask >> 1
		} else {
			break
		}
	}
	return num
}

func isUtf8(data []byte) bool {
	i := 0
	for i < len(data) {
		if (data[i] & 0x80) == 0x00 {
			// 0XXX_XXXX
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			// 110X_XXXX 10XX_XXXX
			// 1110_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_0XXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_10XX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// 1111_110X 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX 10XX_XXXX
			// preNUm() 返回首个字节的8个bits中首个0bit前面1bit的个数，该数量也是该字符所使用的字节数
			i++
			for j := 0; j < num-1; j++ {
				//判断后面的 num - 1 个字节是不是都是10开头
				if (data[i] & 0xc0) != 0x80 {
					return false
				}
				i++
			}
		} else {
			//其他情况说明不是utf-8
			return false
		}
	}
	return true
}
