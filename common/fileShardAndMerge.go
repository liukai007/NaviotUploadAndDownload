package common

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

var uploadDir = "E:\\store"
var downloadDir = "E:\\down"

func ShardFile(filePath string) {
	chunkSize := int64(SliceBytes)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
	}
	fileMetadata := ProduceMetaData(filePath)
	StoreMetadata(filePath, &fileMetadata)

	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))

	fi, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	b := make([]byte, chunkSize)
	var i int64 = 1
	for ; i <= int64(num); i++ {
		ss := (i - 1) * (chunkSize)
		fi.Seek(ss, 0)

		if len(b) > int((fileInfo.Size() - ss)) {
			b = make([]byte, fileInfo.Size()-ss)
		}

		fi.Read(b)
		dirPathStr := uploadDir + "\\" + fileMetadata.Fid
		if !IsDir(dirPathStr) {
			err := os.Mkdir(dirPathStr, 0666)
			if err != nil {
				fmt.Println(err)
			}
		}

		f, err := os.OpenFile(dirPathStr+"\\"+strconv.Itoa(int(i))+".db", os.O_CREATE|os.O_WRONLY, os.ModePerm)
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
		f, err := os.OpenFile(downloadDir+"/"+strconv.Itoa(int(i))+".db", os.O_RDONLY, os.ModePerm)
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
