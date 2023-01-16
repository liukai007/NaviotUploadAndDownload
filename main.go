//package main
//
//import (
//	"fmt"
//	"naviotUploadAndDownload/common"
//)
//
//func main() {
//	fmt.Println(common.FileMD5("C:\\Program Files\\Go\\src\\runtime\\internal\\sys\\zversion.go"))
//	fmt.Println(common.GetFileSize("C:\\Program Files\\Go\\src\\runtime\\internal\\sys\\zversion.go"))
//	fileMetadata := common.ProduceMetaData("C:\\Program Files\\Go\\src\\runtime\\internal\\sys\\zversion.go")
//	fmt.Println(fileMetadata.Fid)
//	fmt.Println(fileMetadata.Md5Sum)
//	fmt.Println(fileMetadata.SliceNum)
//
//	common.StoreMetadata("e:/store/"+fileMetadata.FileName+".slice", &fileMetadata)
//	fileMetadata1, _ := common.LoadMetadata("e:/store/" + fileMetadata.FileName + ".slice")
//
//	fmt.Println("11====" + fileMetadata1.Fid)
//}

//package main
//
//import (
//	"fmt"
//	"io/ioutil"
//	"math"
//	"os"
//	"strconv"
//)
//
//const chunkSize int64 = 4 << 20
//
//func main() {
//	filePath := "E:\\store\\naviot-gateway-1.0-SNAPSHOT.jar"
//	fileInfo, err := os.Stat(filePath)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))
//
//	fi, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	b := make([]byte, chunkSize)
//	var i int64 = 1
//	for ; i <= int64(num); i++ {
//
//		fi.Seek((i-1)*(chunkSize), 0)
//
//		if len(b) > int((fileInfo.Size() - (i-1)*chunkSize)) {
//			b = make([]byte, fileInfo.Size()-(i-1)*chunkSize)
//		}
//
//		fi.Read(b)
//
//		f, err := os.OpenFile("./"+strconv.Itoa(int(i))+".db", os.O_CREATE|os.O_WRONLY, os.ModePerm)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		f.Write(b)
//		f.Close()
//	}
//	fi.Close()
//	fii, err := os.OpenFile(filePath+"1", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	for i := 1; i <= num; i++ {
//		f, err := os.OpenFile("./"+strconv.Itoa(int(i))+".db", os.O_RDONLY, os.ModePerm)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		b, err := ioutil.ReadAll(f)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		fii.Write(b)
//		f.Close()
//	}
//}

package main

import (
	"naviotUploadAndDownload/common"
)

const chunkSize int64 = 4 << 20

func main() {
	filePath := "E:\\store\\naviot-gateway-1.0-SNAPSHOT.jar"
	common.ShardFile(filePath)
	//common.MergeFile(filePath)
	//fmt.Println(common.FileMD5(filePath))
	//fmt.Println(common.FileMD5(filePath + "1"))
}
