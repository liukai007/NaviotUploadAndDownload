package models

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

//验证是否IPv4
func IsIpv4Address(ip string) (bool, error) {
	err := errors.New("NO IPV4")
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false, err
	}
	stringsIp := strings.Split(ip, ".")
	if len(stringsIp) != 4 {
		return false, err
	}
	for i := range stringsIp {
		i3, err := strconv.Atoi(stringsIp[i])
		if err != nil {
			return false, err
		}
		if i3 >= 0 && i3 <= 255 {
			continue
		} else {
			return false, err
		}
	}
	return true, nil
}

//IP转成整数
//测试IP  125.213.100.123  返回值是2111136891
func Ip2Int(ip string) (int64, error) {
	//验证是否是IP地址
	ipBool, err := IsIpv4Address(ip)
	if !ipBool {
		return 0, err
	}
	var result int
	stringsIp := strings.Split(ip, ".")
	for i := range stringsIp {
		i3, _ := strconv.Atoi(stringsIp[i])
		result = i3 | (result << 8)
	}
	return int64(result), nil
}

//整数转成IP
func Int2Ip(ipInt int64) (string, error) {
	v1 := strconv.FormatInt(int64(((ipInt >> 24) & 0xff)), 10)
	v2 := strconv.FormatInt(int64(((ipInt >> 16) & 0xff)), 10)
	v3 := strconv.FormatInt(int64(((ipInt >> 8) & 0xff)), 10)
	v4 := strconv.FormatInt(int64((ipInt & 0xff)), 10)

	return v1 + "." +
		v2 + "." +
		v3 + "." +
		v4, nil
}

//得到所有的IP
//参考 https://www.likecs.com/show-930713.html
func getAllIpAddress(startIp, endIp string) (*list.List, error) {
	allIpList := list.New()
	ipBool, err := IsIpv4Address(startIp)
	if !ipBool {
		return allIpList, err
	}
	ipBool, err = IsIpv4Address(endIp)
	if !ipBool {
		return allIpList, err
	}
	startInt, _ := Ip2Int(startIp)
	endInt, _ := Ip2Int(endIp)
	if endInt <= startInt {
		return allIpList, errors.New("The ipAddress order is wrong")
	}
	for ; startInt <= endInt; startInt++ {
		content, _ := Int2Ip(startInt)
		allIpList.PushBack(content)
	}
	return allIpList, nil
}

//得到本地IP和子网掩码
func LocalIp() (string, string) {
	var finalIp string
	var finalSubnetMask string
	cmd := exec.Command("cmd", "/c", "ipconfig")
	if out, err := cmd.StdoutPipe(); err != nil {
		fmt.Println(err)
	} else {
		defer out.Close()
		if err := cmd.Start(); err != nil {
			fmt.Println(err)
		}

		if opBytes, err := ioutil.ReadAll(out); err != nil {
			log.Fatal(err)
		} else {
			str := string(opBytes)

			var strs = strings.Split(str, "\r\n")

			if 0 != len(strs) {
				var havingFinalIp4 bool = false
				var cnt int = 0
				for index, value := range strs {
					vidx := strings.Index(value, "IPv4")
					//说明已经找到该ip
					if vidx != -1 {
						ip4lines := strings.Split(value, ":")
						if len(ip4lines) == 2 {
							cnt = index
							havingFinalIp4 = true
							ip4str := ip4lines[1]
							finalIp = strings.TrimSpace(ip4str)
						}

					}
					if havingFinalIp4 && index == cnt+1 {
						lindex := strings.Index(value, ":")
						if -1 != lindex {
							lines := strings.Split(value, ":")
							if len(lines) == 2 {
								finalSubnetMask = strings.TrimSpace(lines[1])
							}
						}
					}

					if havingFinalIp4 && index == cnt+2 {
						lindex := strings.Index(value, ":")
						if -1 != lindex {
							lines := strings.Split(value, ":")
							if len(lines) >= 2 {
								fip := lines[1]
								if strings.TrimSpace(fip) != "" {
									break
								}
							}
						}
						havingFinalIp4 = false
						finalIp = ""
					}
				}
			}
		}
	}
	return finalIp, finalSubnetMask
}

//通过主机IP 和子网掩码得到IP网段
//参考: https://blog.csdn.net/oldboy1999/article/details/125782148
//返回值 主机地址的List,网络地址，广播地址，主机个数
func GetIpSegment(ipAddress, subnetMask string) (list1 *list.List, wl string, gb string, hostSum int64, err1 error) {
	//err := errors.New("NO IPV4")
	//验证是否是IP地址
	_, err := IsIpv4Address(ipAddress)
	if err != nil {
		return nil, "", "", 0, err
	}
	_, err = IsIpv4Address(subnetMask)
	if err != nil {
		return nil, "", "", 0, errors.New("NO SubnetMark")
	}
	ipStrings := strings.Split(ipAddress, ".")
	maskStrings := strings.Split(subnetMask, ".")
	//for i := range ipStrings {
	//	fmt.Println(ipStrings[i])
	//}
	//for i := range maskStrings {
	//	fmt.Println(maskStrings[i])
	//}
	//网络地址
	var wl1 string
	//广播地址
	var gb1 string
	var gb1Str string
	for i := range maskStrings {
		i3, _ := strconv.Atoi(maskStrings[i])
		ss := DecimalToBinary(i3)
		if ss == "" {
			ss = "00000000"
		}
		gb1Str = gb1Str + ss
	}
	var j0 int64
	var j1 int64
	for i := range gb1Str {
		if gb1Str[i] == '0' {
			j0++
		} else {
			j1++
		}
	}
	//fmt.Println("有几个0")
	//fmt.Println(j0)
	//fmt.Println("有几个1")
	//fmt.Println(j1)
	for i := 0; i <= 3; i++ {
		a, _ := strconv.Atoi(ipStrings[i])
		b, _ := strconv.Atoi(maskStrings[i])
		//fmt.Println(a & b)
		if j1 >= 24 {
			if i != 3 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10) + "."
				gb1 = gb1 + strconv.FormatInt(int64((a&b)), 10) + "."
			} else {
				sss := int64(a & b)
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10)
				//bin2Str := strconv.FormatInt(int64((a & b)), 2)
				//fmt.Println(bin2Str)
				var o string
				for i := int64(0); i < j0; i++ {
					o = o + "1"
				}
				p, _ := strconv.ParseInt(o, 2, 64)
				//fmt.Println("有几个1111")
				//fmt.Println(p)
				ssss := sss + p
				gb1 = gb1 + strconv.FormatInt(ssss, 10)
			}
		} else if j1 > 16 {
			if i == 0 || i == 1 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10) + "."
				gb1 = gb1 + strconv.FormatInt(int64((a&b)), 10) + "."
			} else if i == 2 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10) + "."
				sss := int64(a & b)
				//bin2Str := strconv.FormatInt(int64((a & b)), 2)
				//fmt.Println(bin2Str)
				var o string
				for i := int64(0); i < j0-8; i++ {
					o = o + "1"
				}
				p, _ := strconv.ParseInt(o, 2, 64)
				//fmt.Println("有几个1111")
				//fmt.Println(p)
				ssss := sss + p
				gb1 = gb1 + strconv.FormatInt(ssss, 10) + "."
			} else if i == 3 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10)
				gb1 = gb1 + "255"
			}
		} else if j1 > 8 {
			if i == 0 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10) + "."
				gb1 = gb1 + strconv.FormatInt(int64((a&b)), 10) + "."
			} else if i == 1 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10) + "."
				sss := int64(a & b)
				//bin2Str := strconv.FormatInt(int64((a & b)), 2)
				//fmt.Println(bin2Str)
				var o string
				for i := int64(0); i < j0-16; i++ {
					o = o + "1"
				}
				p, _ := strconv.ParseInt(o, 2, 64)
				//fmt.Println("有几个1111")
				//fmt.Println(p)
				ssss := sss + p
				gb1 = gb1 + strconv.FormatInt(ssss, 10) + "."
			} else if i == 2 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10) + "."
				gb1 = gb1 + "255" + "."
			} else if i == 3 {
				wl1 = wl1 + strconv.FormatInt(int64((a&b)), 10)
				gb1 = gb1 + "255"
			}

		}

	}

	//广播地址

	//多少个主机数量
	var hostSum1 int64
	hostSum1 = Pow(2, j0) - 2
	l := list.New()
	start1, _ := Ip2Int(wl1)
	end2, _ := Ip2Int(gb1)
	start2Ip, _ := Int2Ip(start1 + 1)
	end2Ip, _ := Int2Ip(end2 - 1)

	allIpList, err := getAllIpAddress(start2Ip, end2Ip)
	if err == nil {
		l = allIpList
	}

	return l, wl1, gb1, hostSum1, nil
}
