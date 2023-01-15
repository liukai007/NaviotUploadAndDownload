package common

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

func ShardFile(filePath string) {
	chunkSize := SliceBytes
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
	}

	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))

	fi, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	b := make([]byte, chunkSize)
	var i int64 = 1
	for ; i <= int64(num); i++ {
		ss := int64(i-1) * int64(chunkSize)
		fi.Seek(ss, 0)

		if len(b) > int((fileInfo.Size() - ss)) {
			b = make([]byte, fileInfo.Size()-ss)
		}

		fi.Read(b)

		f, err := os.OpenFile("./"+strconv.Itoa(int(i))+".db", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		f.Write(b)
		f.Close()
	}
	fi.Close()
}

func MergeFile(filePath string) {
	num := 104
	fii, err := os.OpenFile(filePath+"1", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 1; i <= num; i++ {
		f, err := os.OpenFile("./"+strconv.Itoa(int(i))+".db", os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(err)
			return
		}
		fii.Write(b)
		f.Close()
	}
}
