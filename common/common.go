package common

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileMetadata struct {
	Fid        string    // 操作文件ID，随机生成的UUID
	FileSize   int64     // 文件大小（字节单位）
	FileName   string    // 文件名称
	SliceNum   int64     // 切片数量
	Md5Sum     string    // 文件md5值
	ModifyTime time.Time // 文件修改时间
}

// StoreMetadata 保存文件元数据
func StoreMetadata(metaDataSavePath string, metadata *FileMetadata) error {
	DirCreate(filepath.Dir(metaDataSavePath))
	metaDataSavePath = metaDataSavePath + ".metaData"
	// 写入文件
	file, err := os.OpenFile(metaDataSavePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("写元数据文件%s失败\n", metaDataSavePath)
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(metadata)
	if err != nil {
		fmt.Printf("写元数据文件%s失败\n", metaDataSavePath)
		return err
	}
	return nil
}

// LoadMetadata 加载元数据文件信息
func LoadMetadata(filePath string) (*FileMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("获取文件状态失败，文件路径：", filePath)
		return nil, err
	}

	var metadata FileMetadata
	fileData := gob.NewDecoder(file)
	err = fileData.Decode(&metadata)
	if err != nil {
		fmt.Println("格式化文件元数据失败, err", err)
		return nil, err
	}
	return &metadata, nil
}

func ProduceMetaData(filePath string) FileMetadata {
	uuid, _ := GetUUID()
	md5Value, _ := FileMD5(filePath)
	fileMetadata := FileMetadata{
		Fid:        uuid,
		FileSize:   GetFileSize(filePath),
		FileName:   filepath.Base(filePath),
		SliceNum:   GetSliceNum(GetFileSize(filePath)),
		Md5Sum:     md5Value,
		ModifyTime: time.Now(),
	}
	return fileMetadata
}
