package models

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func HttpGet(url string) error {
	// 请求xx网站首页
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	// 延迟关闭
	defer resp.Body.Close()

	//
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// 请求结果
	fmt.Println(string(body))

	// 请求头
	fmt.Println(resp.Header)

	// 请求相应码
	fmt.Println(resp.Status)

	return nil
}
