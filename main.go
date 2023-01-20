package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	common "fyne/common"
	"fyne/models"
	"github.com/flopp/go-findfont"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var confs = &common.ServiceConfig{}
var globalWait1 sync.WaitGroup

func download(w http.ResponseWriter, request *http.Request) {
	//文件名
	filename := request.FormValue("filename")
	//文件目录
	dirStr := request.FormValue("dirStr")

	//打开文件
	filePath := path.Join(confs.StoreDir, dirStr, filename)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("打开文件%s失败, err:%s\n", filePath, err)
		http.Error(w, "文件打开失败", http.StatusBadRequest)
		return
	}
	//结束后关闭文件
	defer file.Close()

	//设置响应的header头
	w.Header().Add("Content-type", "application/octet-stream")
	w.Header().Add("content-disposition", "attachment; filename=\""+filename+"\"")

	//将文件写至responseBody
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "文件下载失败", http.StatusInternalServerError)
		return
	}
}

func downloadBySlice(w http.ResponseWriter, request *http.Request) {
	filename := request.FormValue("filename")
	sliceIndex := request.FormValue("sliceIndex")

	metadata, err := common.LoadMetadata(common.UploadDir + "\\" + filename + ".metaData")
	if err != nil {
		http.Error(w, "分片文件下载失败", http.StatusBadRequest)
	}

	sliceFile := path.Join(confs.StoreDir, metadata.Fid, sliceIndex)
	if !common.IsFile(sliceFile) {
		fmt.Println("文件切片不存在", sliceFile)
		http.Error(w, "文件异常", http.StatusBadRequest)
		return
	}

	file, err := os.Open(sliceFile)
	if err != nil {
		fmt.Println("打开文件分片失败", sliceFile)
		http.Error(w, "slice read error", http.StatusBadRequest)
		return
	}
	//结束后关闭文件
	defer file.Close()

	//设置响应的header头
	w.Header().Add("Content-type", "application/octet-stream")
	_, err = io.Copy(w, file)
	if err != nil {
		fmt.Printf("下载文件分片%s失败, err:%s", sliceFile, err.Error())
		http.Error(w, "下载文件分片失败", http.StatusBadRequest)
		return
	}
}

// 获取文件元数据信息
func getFileMetaInfo(w http.ResponseWriter, request *http.Request) {
	filename := request.FormValue("filename")
	metaPath := common.UploadDir + "\\" + filename + ".metaData"
	if !common.IsFile(metaPath) {
		fmt.Println("该文件不存在", metaPath)
		http.Error(w, "file not exist", http.StatusBadRequest)
	}

	metadata, err := common.LoadMetadata(metaPath)
	if err != nil {
		http.Error(w, "文件损坏", http.StatusBadRequest)
	}

	cMetadata := metadata
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cMetadata)
	if err != nil {
		fmt.Println("编码文件基本信息失败")
		http.Error(w, "服务异常", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// 记载配置文件
func loadConfig(configPath string) {
	if !common.IsFile(configPath) {
		log.Panicf("config file %s is not exist", configPath)
	}

	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Panicf("load config conf %s failed, err: %s\n", configPath, err)
	}

	err = json.Unmarshal(buf, confs)
	if err != nil {
		log.Panicf("decode config file %s failed, err: %s\n", configPath, err)
	}
}

var tileInfo string
var done = make(chan bool)
var stop = make(chan int, 1)
var num int64
var stopSearch = false

// MainShow 主界面函数
func MainShow(w fyne.Window) {

	title := widget.NewLabel("NAVIoT自发现程序")
	currentIp := widget.NewLabel("本机IP:")
	otherIpAndSubnetMask := widget.NewLabel("其他IP/SubnetMask可以多个，使用逗号隔开:\n例如: 172.168.0.1/255.255.255.0,192.168.1.1/255.255.255.0")
	otherIpAndSubnetMask1 := widget.NewLabel("")
	hello := widget.NewLabel("单文件路径:")
	//models.WriteFile("test1.txt", "{\"a\":\"a\"}")
	//jsonStr := models.GetSmallFileContent("test1.txt")
	//if json.Valid([]byte(jsonStr)) {
	//	fmt.Println("是json字符串")
	//} else {
	//	fmt.Println("不是json字符串")
	//}
	//本机IP文本框
	entryLocalIp := widget.NewEntry()           //文本输入框
	entryOtherIpSubnetMask := widget.NewEntry() //文本输入框
	entry1 := widget.NewEntry()                 //下发文件路径先上传，再下载
	entrySendPath := widget.NewEntry()          //下发文件目录
	labelSendPathName := widget.NewLabel("下发文件目录:")
	//执行命令
	entryExecCmd := widget.NewEntry()
	//执行命令的主机IP 不填写是全部，填写就是固定的主机
	entrySmartHostIp := widget.NewEntry()
	labelExecCmd := widget.NewLabel("执行CMD:")
	labelSmartHostIp := widget.NewLabel("智慧主机IP(空为全,多个逗号隔开):")
	ip, subnetMask := models.LocalIp()
	fmt.Println("子网掩码:" + subnetMask)
	entryLocalIp.SetText(ip)
	dia1 := widget.NewButton("下发文件", func() { //回调函数：打开选择文件对话框
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				log.Println("Cancelled")
				return
			}

			entry1.SetText(reader.URI().Path()) //把读取到的路径显示到输入框中
		}, w)
		fd.Show() //控制是否弹出选择文件目录对话框
	})

	text := widget.NewMultiLineEntry() //多行输入组件
	//text.Disable()                     //禁用输入框，不能更改数据

	labelLast := widget.NewLabel("LK    ALL Right Reserved")
	//labelLast := widget.NewLabel(" ")
	//label4 := widget.NewLabel("多文件路径:")
	//entryFileList := widget.NewEntry()
	//dia2 := widget.NewButton("下发文件夹", func() {
	//	dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
	//		if err != nil {
	//			dialog.ShowError(err, w)
	//			return
	//		}
	//		if list == nil {
	//			log.Println("Cancelled")
	//			return
	//		}
	//		//设置输入框内容
	//		entryFileList.SetText(list.Path())
	//	}, w)
	//})
	//开始搜索按钮
	bt3 := widget.NewButton("开始 搜索", func() {
		stopSearch = false
		list, wl, gb, hostSum, err := models.GetIpSegment(ip, subnetMask)
		if err != nil {
			dialog.ShowError(errors.New("IP错误"), w)
		}
		fmt.Println(list.Len())
		fmt.Println("网络地址是:" + wl)
		fmt.Println("广播地址是:" + gb)
		fmt.Println("主机个数是:" + strconv.FormatInt(int64(hostSum), 10))
		var wg2 sync.WaitGroup
		for i := list.Front(); i != nil; i = i.Next() {
			if stopSearch {
				fmt.Println("停止搜索")
				break
			}
			wg2.Add(1)
			//fmt.Printf("item = %v\n", i.Value)
			ipAddressTmp := fmt.Sprint(i.Value)

			go func() {
				err := models.HttpGet("http://" + ipAddressTmp + ":27777/ping")
				if err != nil {
					fmt.Println(ipAddressTmp + " 链接异常")
				} else {
					fmt.Println(ipAddressTmp + " 链接成功")
					models.WriteFileAppend("hostIpAddress.txt", ipAddressTmp+"\n")
				}
				wg2.Done()
			}()
		}
		wg2.Wait()
		//其他IP和subnetMark
		var wg1 sync.WaitGroup
		tmpText := entryOtherIpSubnetMask.Text
		tmpText = strings.TrimSpace(tmpText)
		if tmpText != "" {
			strings1 := strings.Split(tmpText, ",")

			for j := range strings1 {
				str1 := strings1[j]
				str2 := strings.Split(str1, "/")
				if len(str2) != 2 {
					fmt.Println("缺少IP或者子网掩码")
					continue
				}
				list1, _, _, _, err1 := models.GetIpSegment(str2[0], str2[1])
				if err1 != nil {
					dialog.ShowError(errors.New("IP错误"), w)
				}
				for i := list1.Front(); i != nil; i = i.Next() {
					if stopSearch {
						fmt.Println("停止搜索")
						break
					}
					wg1.Add(1)
					//fmt.Printf("item = %v\n", i.Value)
					ipAddressTmp := fmt.Sprint(i.Value)

					go func() {
						err := models.HttpGet("http://" + ipAddressTmp + ":27777/ping")
						if err != nil {
							fmt.Println(ipAddressTmp + " 链接异常")
						} else {
							fmt.Println(ipAddressTmp + " 链接成功")
							models.WriteFileAppend("hostIpAddress.txt", ipAddressTmp+"\n")
						}
						wg1.Done()
					}()
				}
			}
		}

		wg1.Wait()
		multiLine := widget.NewMultiLineEntry()
		jsonStr := models.GetSmallFileContent("hostIpAddress.txt")
		multiLine.SetText(jsonStr)
		content := container.NewVBox(
			multiLine,
		)
		cd := dialog.NewCustom("获取的主机IP", "dismiss", content, w)
		cd.Resize(fyne.NewSize(300, 170))
		cd.SetDismissText("关闭")
		cd.Show()

	})

	bt5 := widget.NewButton("重新 搜索", func() {
		stopSearch = false
		models.RemoveFile("hostIpAddress.txt")
		list, wl, gb, hostSum, err := models.GetIpSegment(ip, subnetMask)
		if err != nil {
			dialog.ShowError(errors.New("IP错误"), w)
		}
		fmt.Println(list.Len())
		fmt.Println("网络地址是:" + wl)
		fmt.Println("广播地址是:" + gb)
		fmt.Println("主机个数是:" + strconv.FormatInt(int64(hostSum), 10))
		var wg sync.WaitGroup
		for i := list.Front(); i != nil; i = i.Next() {
			if stopSearch {
				fmt.Println("停止搜索")
				break
			}
			wg.Add(1)
			//fmt.Printf("item = %v\n", i.Value)
			ipAddressTmp := fmt.Sprint(i.Value)

			go func() {
				err := models.HttpGet("http://" + ipAddressTmp + ":27777/ping")
				if err != nil {
					fmt.Println(ipAddressTmp + " 链接异常")
				} else {
					fmt.Println(ipAddressTmp + " 链接成功")
					models.WriteFileAppend("hostIpAddress.txt", ipAddressTmp+"\n")
				}
				wg.Done()
			}()
		}
		wg.Wait()

		//其他IP和subnetMark
		var wg1 sync.WaitGroup
		tmpText := entryOtherIpSubnetMask.Text
		tmpText = strings.TrimSpace(tmpText)
		if tmpText != "" {
			strings1 := strings.Split(tmpText, ",")

			for j := range strings1 {
				str1 := strings1[j]
				str2 := strings.Split(str1, "/")
				if len(str2) != 2 {
					fmt.Println("缺少IP或者子网掩码")
					continue
				}
				list1, _, _, _, err1 := models.GetIpSegment(str2[0], str2[1])
				if err1 != nil {
					dialog.ShowError(errors.New("IP错误"), w)
				}
				for i := list1.Front(); i != nil; i = i.Next() {
					if stopSearch {
						fmt.Println("停止搜索")
						break
					}
					wg1.Add(1)
					//fmt.Printf("item = %v\n", i.Value)
					ipAddressTmp := fmt.Sprint(i.Value)

					go func() {
						err := models.HttpGet("http://" + ipAddressTmp + ":27777/ping")
						if err != nil {
							fmt.Println(ipAddressTmp + " 链接异常")
						} else {
							fmt.Println(ipAddressTmp + " 链接成功")
							models.WriteFileAppend("hostIpAddress.txt", ipAddressTmp+"\n")
						}
						wg1.Done()
					}()
				}
			}
		}

		wg1.Wait()
		multiLine := widget.NewMultiLineEntry()
		jsonStr := models.GetSmallFileContent("hostIpAddress.txt")
		multiLine.SetText(jsonStr)
		content := container.NewVBox(
			multiLine,
		)
		cd := dialog.NewCustom("获取的主机IP", "dismiss", content, w)
		cd.Resize(fyne.NewSize(300, 170))
		cd.SetDismissText("关闭")
		cd.Show()

	})

	head := container.NewCenter(title)

	v0 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), currentIp, layout.NewSpacer(), entryLocalIp)
	v01 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), otherIpAndSubnetMask, layout.NewSpacer())
	v02 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), otherIpAndSubnetMask1, layout.NewSpacer(), entryOtherIpSubnetMask)
	v1 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), hello, dia1, entry1)
	//SendPathNameSetting := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), labelSendPathName, entrySendPath)
	SendPathNameSetting := container.NewBorder(nil, nil, labelSendPathName, layout.NewSpacer(), entrySendPath)
	//v4 := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), label4, dia2, entryFileList)
	execCmdContent := container.NewBorder(nil, nil, labelExecCmd, layout.NewSpacer(), entryExecCmd)
	smartHostIp := container.NewBorder(nil, nil, labelSmartHostIp, layout.NewSpacer())
	smartHostIp1 := container.NewBorder(nil, nil, nil, layout.NewSpacer(), entrySmartHostIp)

	v5 := container.NewHBox(bt3, bt5)
	v5Center := container.NewCenter(v5)

	/****上传 start**************/
	//上传 一个上传按钮
	uploaderFilePath := widget.NewLabel("上传文件:")
	uploaderFileEntry := widget.NewEntry()
	uploaderFileDia1 := widget.NewButton("查找文件", func() { //回调函数：打开选择文件对话框
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				log.Println("Cancelled")
				return
			}

			uploaderFileEntry.SetText(reader.URI().Path()) //把读取到的路径显示到输入框中
		}, w)

		//fd.SetFilter(storage.NewExtensionFileFilter([]string{".txt"})) //打开的文件格式类型
		fd.Show() //控制是否弹出选择文件目录对话框
	})
	uploaderFilePathAndBtn := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), uploaderFilePath, uploaderFileDia1, uploaderFileEntry)
	//是否覆盖已经上传过的
	//https://www.bilibili.com/video/BV1Hy4y1g7iv
	isAgain := false
	chk1 := widget.NewCheck("覆盖上传", nil)
	chk1.OnChanged = func(b bool) {
		if chk1.Checked {
			isAgain = true
		} else {
			isAgain = false
		}
	}
	//上传按钮
	uploaderBt3 := widget.NewButton("上传", func() {
		content := strings.TrimSpace(uploaderFileEntry.Text)
		if content == "" {
			fmt.Println("上传文件路径不能为空")
			//需要一个弹框
			multiLine := widget.NewMultiLineEntry()
			multiLine.SetText("上传路径为空")
			content1 := container.NewVBox(
				multiLine,
			)
			cd := dialog.NewCustom("提醒", "dismiss", content1, w)
			cd.Resize(fyne.NewSize(170, 170))
			cd.SetDismissText("关闭")
			cd.Show()
			return
		}
		common.ShardFile(content, isAgain)
		//需要一个弹框
		multiLine := widget.NewMultiLineEntry()
		multiLine.SetText("上传完毕")
		content1 := container.NewVBox(
			multiLine,
		)
		cd := dialog.NewCustom("提醒", "dismiss", content1, w)
		cd.Resize(fyne.NewSize(170, 170))
		cd.SetDismissText("关闭")
		cd.Show()
	})
	uploaderBox := container.NewHBox(chk1, uploaderBt3)
	uploaderBoxCenter := container.NewCenter(uploaderBox)

	/****上传 end**************/

	/****下发 start**************/
	isAgainDownload := false
	chk2 := widget.NewCheck("覆盖下发", nil)
	chk2.OnChanged = func(b bool) {
		if chk2.Checked {
			isAgainDownload = true
		} else {
			isAgainDownload = false
		}
	}

	downloaderBt := widget.NewButton("下发", func() {
		fmt.Println("检测是否已经上传过,没有上传就先上传")
		content := strings.TrimSpace(entry1.Text)
		if content == "" {
			fmt.Println("下发文件路径不能为空")
			//需要一个弹框
			multiLine := widget.NewMultiLineEntry()
			multiLine.SetText("下发路径为空")
			content1 := container.NewVBox(
				multiLine,
			)
			cd := dialog.NewCustom("提醒", "dismiss", content1, w)
			cd.Resize(fyne.NewSize(170, 170))
			cd.SetDismissText("关闭")
			cd.Show()
			return
		}
		common.ShardFile(content, isAgainDownload)

		fmt.Println("开始下发,目前下发的IP有")
		jsonStr := models.GetSmallFileContent("hostIpAddress.txt")
		jsonStr = strings.TrimSpace(jsonStr)
		if jsonStr == "" {
			fmt.Println("没有需要下发的主机")
			//需要一个弹框
			multiLine := widget.NewMultiLineEntry()
			multiLine.SetText("没有下发的主机")
			content1 := container.NewVBox(
				multiLine,
			)
			cd := dialog.NewCustom("提醒", "dismiss", content1, w)
			cd.Resize(fyne.NewSize(170, 170))
			cd.SetDismissText("关闭")
			cd.Show()
			return
		}
		strs := strings.Split(jsonStr, "\n")
		var a map[string]string
		a = make(map[string]string)
		for i := range strs {
			a[strs[i]] = "1"
		}
		for s := range a {
			fmt.Println(s)
			//http://127.0.0.1:7777/downloadFile?filename=ubuntu-18.04.4-desktop-amd64.iso&downloadDir=E:\down
			sendFileName := path.Base(entry1.Text)
			sendFilePath := entrySendPath.Text
			go func(s string) {
				err := models.HttpGet("http://" + s + ":27777/downloadFile?filename=" +
					sendFileName + "&downloadDir=" + sendFilePath)
				if err != nil {
					fmt.Println(s + " 链接异常")
				} else {
					fmt.Println(s + " 链接成功")
					time.Sleep(1000)
				}
			}(s)
		}
	})
	//下发成功日志
	verifyDownloaderFileBt := widget.NewButton("下发日志查看", func() {
		fmt.Println("下发日志生成中")
		content := strings.TrimSpace(entry1.Text)
		if content == "" {
			fmt.Println("下发日志生成失败，下发文件路径为空")
			//需要一个弹框
			multiLine := widget.NewMultiLineEntry()
			multiLine.SetText("下发路径为空")
			content1 := container.NewVBox(
				multiLine,
			)
			cd := dialog.NewCustom("提醒", "dismiss", content1, w)
			cd.Resize(fyne.NewSize(170, 170))
			cd.SetDismissText("关闭")
			cd.Show()
			return
		}
		fmt.Println("开始下发,目前下发的IP有")
		jsonStr := models.GetSmallFileContent("hostIpAddress.txt")
		jsonStr = strings.TrimSpace(jsonStr)
		if jsonStr == "" {
			fmt.Println("下发日志生成失败，没有需要下发的主机")
			return
		}
		strs := strings.Split(jsonStr, "\n")
		var a map[string]string
		a = make(map[string]string)
		for i := range strs {
			a[strs[i]] = "1"
		}
		sendFileName := path.Base(entry1.Text)
		sendFilePath := entrySendPath.Text
		if common.IsFile("下发日志=" + sendFileName + ".log") {
			models.RemoveFile("下发日志=" + sendFileName + ".log")
		}
		f, err := os.Create("下发日志=" + sendFileName + ".log")
		if err != nil {
			//panic(err)
		}
		wr := &SyncWriter{sync.Mutex{}, f}
		wg := sync.WaitGroup{}
		for s := range a {
			wg.Add(1)
			go func(s string) {
				fmt.Println(s)
				statusCode, content, err := models.HttpGetGetValue("http://" + s + ":27777/verifyDownloadFile?filename=" +
					sendFileName + "&downloadDir=" + sendFilePath)
				if err != nil {
					fmt.Fprintln(wr, sendFileName+" "+" 发送失败 主机("+s+")目录("+sendFilePath+")\n")
					wg.Done()
					return
				}
				if statusCode == 200 {
					if content == "true" {
						fmt.Fprintln(wr, sendFileName+" "+" 发送成功 主机("+s+")目录("+sendFilePath+")\n")
						wg.Done()
					} else {
						fmt.Fprintln(wr, sendFileName+" "+" 发送失败 主机("+s+")目录("+sendFilePath+")\n")
						wg.Done()
					}
				} else {
					fmt.Fprintln(wr, sendFileName+" "+" 发送失败 主机("+s+")目录("+sendFilePath+")\n")
					wg.Done()
				}
			}(s)
		}
		wg.Wait()
		//需要一个弹框
		multiLine := widget.NewMultiLineEntry()
		multiLine.SetText("下发日志生成完毕")
		content1 := container.NewVBox(
			multiLine,
		)
		cd := dialog.NewCustom("提醒", "dismiss", content1, w)
		cd.Resize(fyne.NewSize(170, 170))
		cd.SetDismissText("关闭")
		cd.Show()
	})

	downloaderBox := container.NewHBox(chk2, downloaderBt, verifyDownloaderFileBt)
	downloaderBoxCenter := container.NewCenter(downloaderBox)
	/****下发 end**************/

	/****空白 end**************/
	blank := container.NewCenter()
	/****空白 end**************/

	/** 执行命令 start****/
	execCmdBtn := widget.NewButton("执行命令", func() {
		if strings.TrimSpace(entryExecCmd.Text) == "" {
			fmt.Println("执行命令内容不能空")
			return
		}
		fmt.Println("开始执行命令,目前执行的IP有")
		ipStrings := entrySmartHostIp.Text
		if ipStrings != "" {
			ipList := strings.Split(ipStrings, ",")
			for i := range ipList {
				s := ipList[i]
				fmt.Println("执行IP: " + s)
				globalWait1.Add(1)
				go execCmdPost(s, entryExecCmd.Text)
			}
			globalWait1.Wait()
			return
		}
		jsonStr := models.GetSmallFileContent("hostIpAddress.txt")
		jsonStr = strings.TrimSpace(jsonStr)
		if jsonStr == "" {
			fmt.Println("没有需要下发的主机")
			//需要一个弹框
			multiLine := widget.NewMultiLineEntry()
			multiLine.SetText("没有下发的主机")
			content1 := container.NewVBox(
				multiLine,
			)
			cd := dialog.NewCustom("提醒", "dismiss", content1, w)
			cd.Resize(fyne.NewSize(170, 170))
			cd.SetDismissText("关闭")
			cd.Show()
			return
		}
		strs := strings.Split(jsonStr, "\n")
		var a map[string]string
		a = make(map[string]string)
		for i := range strs {
			a[strs[i]] = "1"
		}
		for s := range a {
			globalWait1.Add(1)
			go execCmdPost(s, entryExecCmd.Text)
		}
		globalWait1.Wait()
	})
	execCmdBox := container.NewVBox(execCmdBtn)
	execCmdBoxBoxCenter := container.NewCenter(execCmdBox)
	/** 执行命令   end****/

	ctnt := container.NewVBox(head, v0, v01, v02, v5Center,
		uploaderFilePathAndBtn, uploaderBoxCenter,
		blank,
		blank,
		v1,
		SendPathNameSetting,
		downloaderBoxCenter,
		//v4,
		execCmdContent,
		smartHostIp,
		smartHostIp1,
		execCmdBoxBoxCenter,
		text, labelLast) //控制显示位置顺序
	w.SetContent(ctnt)
}

//设置字体
func init() {
	fontPaths := findfont.List()
	for _, fontPath := range fontPaths {
		//fmt.Println(fontPath)
		//楷体:simkai.ttf
		//黑体:simhei.ttf
		//微软雅黑：msyh.ttc
		if strings.Contains(fontPath, "simkai.ttf") {
			err := os.Setenv("FYNE_FONT", fontPath)
			if err != nil {
				return
			}
			break
		}
	}
}

func main() {
	var configPath *string
	if common.IsFile("config.json") {
		configPath = flag.String("configPath", "config.json", "服务配置文件")
	} else {
		models.WriteFile("config.json", "{\"Port\":800,\"Address\":\"0.0.0.0\",\"StoreDir\":\"E:/store\"}")
		configPath = flag.String("configPath", "config.json", "服务配置文件")
	}
	//新建一个app
	a := app.New()
	//设置窗口栏，任务栏图标
	//a.SetIcon(resourceIconPng)
	//新建一个窗口
	w := a.NewWindow("NAVIoT自发现程序V1.0")
	//主界面框架布局
	MainShow(w)
	//尺寸
	w.Resize(fyne.Size{Width: 500, Height: 200})
	//w居中显示
	w.CenterOnScreen()
	go func() {
		/**-----------------------------------------------*/
		flag.Parse()
		loadConfig(*configPath)
		if !common.IsDir(confs.StoreDir) {
			fmt.Println("目录不存在:" + confs.StoreDir)
			os.Mkdir(confs.StoreDir, 0777)
		}
		http.HandleFunc("/getFileMetaInfo", getFileMetaInfo)
		http.HandleFunc("/download", download)
		http.HandleFunc("/downloadBySlice", downloadBySlice)
		err11 := http.ListenAndServe(":"+strconv.Itoa(confs.Port), nil)
		if err11 != nil {
			log.Fatal("ListenAndServe: ", err11)
		}
		/**----------------------------------------------*/
	}()

	//循环运行
	w.ShowAndRun()
	err := os.Unsetenv("FYNE_FONT")
	if err != nil {
		return
	}

}

func execCmd(ipAddr string, cmdContent string) {
	defer globalWait1.Done()
	resp, err := http.Get("http://" + ipAddr + ":27777/execCmd?cmdContent=" + cmdContent)
	if err != nil {
		fmt.Println(ipAddr + " 发送命令异常")
	} else {
		fmt.Println(ipAddr + " 发送命令成功")
		// 延迟关闭
		defer resp.Body.Close()
		//
		body, _ := ioutil.ReadAll(resp.Body)
		// 请求结果
		fmt.Println(string(body))
		if json.Valid([]byte(string(body))) {
			var dat map[string]interface{}
			if err := json.Unmarshal(body, &dat); err == nil {
				fmt.Println(ipAddr)
				fmt.Println(dat["result"])
			}
		} else {
			fmt.Println("不是json字符串")
		}
	}
}

func execCmdPost(ipAddr string, cmdContent string) {
	defer globalWait1.Done()
	config := map[string]interface{}{}
	config["cmdContent"] = cmdContent
	configData, _ := json.Marshal(config)
	fmt.Println(config)
	body := bytes.NewBuffer([]byte(configData))
	resp, err := http.Post("http://"+ipAddr+":27777/execCmd", "application/json;charset=utf-8", body)
	if err != nil {
		fmt.Println(ipAddr + " 发送命令异常")
	} else {
		fmt.Println(ipAddr + " 发送命令成功")
		// 延迟关闭
		defer resp.Body.Close()
		//
		body, _ := ioutil.ReadAll(resp.Body)
		// 请求结果
		fmt.Println(string(body))
		if json.Valid([]byte(string(body))) {
			var dat map[string]interface{}
			if err := json.Unmarshal(body, &dat); err == nil {
				fmt.Println(ipAddr)
				fmt.Println(dat["result"])
			}
		} else {
			fmt.Println("不是json字符串")
		}
	}
}

//多线程写入数据
type SyncWriter struct {
	m      sync.Mutex
	Writer io.Writer
}

func (w *SyncWriter) Write(b []byte) (n int, err error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.Writer.Write(b)
}
