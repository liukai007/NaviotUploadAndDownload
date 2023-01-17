package common

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

var UploadDir = "E:\\store"
var DownloadDir = "E:\\down"

func ShardFile(filePathStr string, isAgain bool) {
	//是否重新上传
	if isAgain {
		fmt.Println("重新上传删除之前上传的文件")
		fileMetadata1, err := LoadMetadata(UploadDir + "\\" + filepath.Base(filePathStr) + ".metaData")
		if err == nil {
			DelFileDir(UploadDir + "\\" + fileMetadata1.Fid)
			DelFile(UploadDir + "\\" + filepath.Base(filePathStr) + ".metaData")
		}
	}
	//判断文件是否已经上传过
	base1 := filepath.Base(filePathStr)
	if IsFile(UploadDir + "\\" + base1 + ".metaData") {
		fmt.Println("同名文件已经上传过,不用再上传了,也不需要校验了")
		return
	}

	chunkSize := int64(SliceBytes)
	fileInfo, err := os.Stat(filePathStr)
	if err != nil {
		fmt.Println(err)
	}
	fileMetadata := ProduceMetaData(filePathStr)
	StoreMetadata(UploadDir+"\\"+filepath.Base(filePathStr), &fileMetadata)

	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))

	fi, err := os.OpenFile(filePathStr, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	b := make([]byte, chunkSize)
	var i int64 = 1
	fmt.Println("分片开始上传")
	for ; i <= int64(num); i++ {
		ss := (i - 1) * (chunkSize)
		fi.Seek(ss, 0)

		if len(b) > int((fileInfo.Size() - ss)) {
			b = make([]byte, fileInfo.Size()-ss)
		}

		fi.Read(b)
		dirPathStr := UploadDir + "\\" + fileMetadata.Fid
		if !IsDir(dirPathStr) {
			err := os.Mkdir(dirPathStr, 0666)
			if err != nil {
				fmt.Println(err)
			}
		}

		f, err := os.OpenFile(dirPathStr+"\\"+strconv.Itoa(int(i))+".db", os.O_CREATE|os.O_WRONLY, os.ModePerm)
		fmt.Println(strconv.Itoa(int(i)) + ".db" + "分片上传中")
		if err != nil {
			fmt.Println(err)
			return
		}
		f.Write(b)
		f.Close()
	}
	fmt.Println("分片结束上传")
	fi.Close()
	//合并校验，校验之后删除合并文件，如果校验失败重新上传，重复最多3次
	MergeFile(UploadDir+"\\"+filepath.Base(filePathStr), UploadDir+"\\"+fileMetadata.Fid, fileMetadata)
	//校验
	verifyMD5Bool := VerifyFileMD5(fileMetadata, UploadDir+"\\"+filepath.Base(filePathStr))
	DelFile(UploadDir + "\\" + filepath.Base(filePathStr))
	if !verifyMD5Bool {
		DelFile(UploadDir + "\\" + filepath.Base(filePathStr) + ".metaData")
		DelFileDir(UploadDir + "\\" + fileMetadata.Fid)
	}
}

//第一个参数是生成文件的目录，第二个参数是分片所在目录,第三个参数是文件元数据
func MergeFile(filePath string, shardPath string, metadata FileMetadata) {
	fmt.Println("文件合并开始")
	num := metadata.SliceNum
	fii, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	defer fii.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 1; i <= int(num); i++ {
		//f, err := os.OpenFile(DownloadDir+"/"+strconv.Itoa(int(i))+".db", os.O_RDONLY, os.ModePerm)
		f, err := os.OpenFile(shardPath+"/"+strconv.Itoa(int(i))+".db", os.O_RDONLY, os.ModePerm)
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
	fmt.Println("文件合并结束")
}

//校验MD5
func VerifyFileMD5(metadata FileMetadata, filepathStr string) bool {
	fmt.Println("文件校验开始")
	defer fmt.Println("文件校验结束")
	md5Str, err := FileMD5(filepathStr)
	if err != nil {
		fmt.Println("MD5生成报错")
		fmt.Println("校验失败,请再次重新上传(MD5生成报错)")
		return false
	}
	if md5Str == metadata.Md5Sum {
		fmt.Println("校验成功")
		return true
	}
	fmt.Println("校验失败,请再次重新上传")
	return false
}
