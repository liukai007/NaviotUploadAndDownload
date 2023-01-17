package models

import "fmt"

//10进制转成2进制
func DecimalToBinary(num int) string {
	var binary []int

	for num != 0 {
		binary = append(binary, num%2)
		num = num / 2
	}
	if len(binary) == 0 {
		fmt.Printf("%d\n", 0)
	} else {
		var s string
		for i := len(binary) - 1; i >= 0; i-- {
			s1 := fmt.Sprintf("%d", binary[i])
			s = s + s1
		}
		return s
	}
	return ""
}
