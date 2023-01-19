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

//返回值 请求码，返回结果，是否报错
func HttpGetGetValue(url string) (int, string, error) {
	// 请求xx网站首页
	resp, err := http.Get(url)

	if err != nil {
		return resp.StatusCode, "", err
	}

	// 延迟关闭
	defer resp.Body.Close()

	//
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, string(body), err
	}
	// 请求结果
	fmt.Println(string(body))

	// 请求头
	fmt.Println(resp.Header)

	// 请求相应码
	fmt.Println(resp.Status)

	return resp.StatusCode, string(body), nil
}
