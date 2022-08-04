package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

func randomSortArray(arr []string) ([]string, error) {
	randarr := make([]string, len(arr))
	copy(randarr, arr)
	sort.Sort(sort.Reverse(sort.StringSlice(randarr)))

	seed := time.Now().UnixNano()
	source := rand.NewSource(int64(seed))
	r := rand.New(source)
	fmt.Println("Rand int:", r.Int63()%int64(len(arr)))

	for index, value := range randarr {
		j := r.Int63() % int64(len(arr))
		temp := value
		randarr[index] = randarr[j]
		randarr[j] = temp
	}
	fmt.Println("arr: ", randarr)
	return randarr, nil
}

func main() {

	keys := strings.Split("abcdefghijklmnopqrstuvwxyz0123456789", "")
	keysLength := len(keys)
	arr1, _ := randomSortArray(keys)
	arr2, _ := randomSortArray(arr1)
	arr3, _ := randomSortArray(arr2)
	arr4, _ := randomSortArray(arr3)

	powKeylen := keysLength * keysLength * keysLength * keysLength
	temp := make([]string, powKeylen)
	var counts int //计数

	//憨包写法，哈哈哈哈哈，管他的，能跑就行！
	file, err := os.OpenFile("./pwd1.txt", os.O_WRONLY, 0666)
	file1, err := os.OpenFile("./pwd2.txt", os.O_WRONLY, 0666)
	file2, err := os.OpenFile("./pwd3.txt", os.O_WRONLY, 0666)
	file3, err := os.OpenFile("./pwd4.txt", os.O_WRONLY, 0666)
	file4, err := os.OpenFile("./pwd5.txt", os.O_WRONLY, 0666)
	file5, err := os.OpenFile("./pwd6.txt", os.O_WRONLY, 0666)
	file6, err := os.OpenFile("./pwd7.txt", os.O_WRONLY, 0666)
	file7, err := os.OpenFile("./pwd8.txt", os.O_WRONLY, 0666)
	file8, err := os.OpenFile("./pwd9.txt", os.O_WRONLY, 0666)
	file9, err := os.OpenFile("./pwd10.txt", os.O_WRONLY, 0666)
	file10, err := os.OpenFile("./pwd11.txt", os.O_WRONLY, 0666)
	file11, err := os.OpenFile("./pwd12.txt", os.O_WRONLY, 0666)
	file12, err := os.OpenFile("./pwd13.txt", os.O_WRONLY, 0666)
	file13, err := os.OpenFile("./pwd14.txt", os.O_WRONLY, 0666)
	file14, err := os.OpenFile("./pwd15.txt", os.O_WRONLY, 0666)
	file15, err := os.OpenFile("./pwd16.txt", os.O_WRONLY, 0666)
	if !os.IsExist(err) {
		fmt.Printf("打开文件失败,%v,正在创建文件...", err)
		file, err = os.Create("./pwd1.txt")
		file1, err = os.Create("./pwd2.txt")
		file2, err = os.Create("./pwd3.txt")
		file3, err = os.Create("./pwd4.txt")
		file4, err = os.Create("./pwd5.txt")
		file5, err = os.Create("./pwd6.txt")
		file6, err = os.Create("./pwd7.txt")
		file7, err = os.Create("./pwd8.txt")
		file8, err = os.Create("./pwd9.txt")
		file9, err = os.Create("./pwd10.txt")
		file10, err = os.Create("./pwd11.txt")
		file11, err = os.Create("./pwd12.txt")
		file12, err = os.Create("./pwd13.txt")
		file13, err = os.Create("./pwd14.txt")
		file14, err = os.Create("./pwd15.txt")
		file15, err = os.Create("./pwd16.txt")

		if err != nil {
			fmt.Printf("创建文件失败！")
			return
		}
		fmt.Println("创建成功！")
	}

	fmt.Println("开始生成密码")

	var j = 0
	for _, key1 := range arr1 {

		for _, key2 := range arr2 {

			for _, key3 := range arr3 {

				for _, key4 := range arr4 {

					secret := fmt.Sprintln(key1 + key2 + key3 + key4)
					temp[j] = secret
					j += 1
					if j == powKeylen {
						// fmt.Println("开始存入文档")
						temp, _ = randomSortArray(temp)
						for _, value := range temp {
							if counts == 0 {
								file.WriteString(value)
								counts++
								continue
							} else if counts == 1 {
								file1.WriteString(value)
								counts++
								continue
							} else if counts == 2 {
								file2.WriteString(value)
								counts++
								continue
							} else if counts == 3 {
								file3.WriteString(value)
								counts++
								continue
							} else if counts == 4 {
								file4.WriteString(value)
								counts++
								continue
							} else if counts == 5 {
								file5.WriteString(value)
								counts++
								continue
							} else if counts == 6 {
								file6.WriteString(value)
								counts++
								continue
							} else if counts == 7 {
								file7.WriteString(value)
								counts++
								continue
							} else if counts == 8 {
								file8.WriteString(value)
								counts++
								continue
							} else if counts == 9 {
								file9.WriteString(value)
								counts++
								continue
							} else if counts == 10 {
								file10.WriteString(value)
								counts++
								continue
							} else if counts == 11 {
								file11.WriteString(value)
								counts++
								continue
							} else if counts == 12 {
								file12.WriteString(value)
								counts++
								continue
							} else if counts == 13 {
								file13.WriteString(value)
								counts++
								continue
							} else if counts == 14 {
								file14.WriteString(value)
								counts++
								continue
							} else if counts == 15 {
								file15.WriteString(value)
								counts = 0
								continue
							}
						}
						j = 0
					}

				}
			}
		}
	}

	defer func(file *os.File) {
		file.Close()
		file1.Close()
		file2.Close()
		file3.Close()
		file4.Close()
		file5.Close()
		file6.Close()
		file7.Close()
		file8.Close()
		file9.Close()
		file10.Close()
		file11.Close()
		file12.Close()
		file13.Close()
		file14.Close()
		file15.Close()
		fmt.Println("结束")
	}(file)

}
