package main

import (
	"fmt"
	"naviotUploadAndDownload/common"
)

func main() {
	fmt.Println(common.FileMD5("C:\\Program Files\\Go\\src\\runtime\\internal\\sys\\zversion.go"))
	fmt.Println(common.GetFileSize("C:\\Program Files\\Go\\src\\runtime\\internal\\sys\\zversion.go"))
	fileMetadata := common.ProduceMetaData("C:\\Program Files\\Go\\src\\runtime\\internal\\sys\\zversion.go")
	fmt.Println(fileMetadata.Fid)
	fmt.Println(fileMetadata.Md5Sum)

	common.StoreMetadata("g:/123", &fileMetadata)
	fileMetadata1, _ := common.LoadMetadata("g:/123")

	fmt.Println("11====" + fileMetadata1.Fid)
}
