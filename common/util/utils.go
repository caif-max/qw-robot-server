package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"qw-robot-server/common/config"
	"qw-robot-server/common/log"
	"regexp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var bArray = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
var sArray = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

/**
 * @Author      : LiuPf
 * @Description : 返回一个长度为@length的随机字母组合
 * @Date        : 2019-03-27 20:04
 * @Modify      :
 */
func RandomLetter(length int) string {
	var randomLetter string
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		randomInt := r.Intn(26)
		if randomInt%2 == 0 {
			randomLetter += bArray[randomInt]
		} else {
			randomLetter += sArray[randomInt]
		}
	}
	return randomLetter
}

/**
 * @Author      : LiuPf
 * @Description : 返回一个长度为@length的随机数字
 * @Date        : 2019-03-27 20:04
 * @Modify      :
 */
func RandomNum(length int) string {
	var randomNum string
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		randomInt := r.Intn(10)
		randomNum += strconv.Itoa(randomInt)
	}
	return randomNum
}

/**
 * @Author      : LiuPf
 * @Description : 返回一个长度为@length的随机数字字母组合
 * @Date        : 2019-03-27 20:04
 * @Modify      :
 */
func RandomLetterAndNum(length int) string {
	var randomLetterAndNum string
	var randomLetter = make([]string, 0)
	var randomNum = make([]string, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		randomInt := r.Intn(26)
		if randomInt%2 == 0 {
			randomLetter = append(randomLetter, bArray[randomInt])
		} else {
			randomLetter = append(randomLetter, sArray[randomInt])
		}
	}
	for i := 0; i < length; i++ {
		randomInt := r.Intn(10)
		randomNum = append(randomNum, strconv.Itoa(randomInt))
	}
	for i := 0; i < length; i++ {
		randomInt := r.Intn(length)
		if randomInt%2 == 0 && randomInt <= len(randomLetter) {
			randomLetterAndNum += randomLetter[i]
		}
		if randomInt%2 != 0 && randomInt <= len(randomNum) {
			randomLetterAndNum += randomNum[i]
		}
	}
	return randomLetterAndNum
}

func IfExistMap(data map[string]interface{}, key string) bool {
	if data[key] != nil {
		return true
	}
	return false
}

// 判断key是否存在
func MapContains(src map[string]bson.M, key string) bool {
	if _, ok := src[key]; ok {
		return true
	}
	return false
}

func MapMerge(ms ...map[string]string) map[string]string {
	res := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}
func ContainsArray(data []interface{}, value interface{}) bool {
	for _, v := range data {
		if v == value {
			return true
		}
	}
	return false
}

func ContainsArrayString(data []string, value string) bool {
	for _, v := range data {
		if v == value {
			return true
		}
	}
	return false
}

func ArrayRemove(slice []string, elems ...string) []string {
loop:
	for i := 0; i < len(slice); i++ {
		url := slice[i]
		for _, rem := range elems {
			if url == rem {
				slice = append(slice[:i], slice[i+1:]...)
				i-- // Important: decrease index
				continue loop
			}
		}
	}
	return slice
}

// 深度拷贝
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	} else if valueMap, ok := value.(bson.M); ok {
		newMap := make(bson.M)
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}
	}
	return value
}

/**
* 结构体转map,并去掉值为空字符串的key，去掉分页字段(limit,skip)
* @param   obj {interface{}} 结构体参数
* @returns d   {map[string]interface{}} 转换后的map
* @returns err {error} 	错误
 */
func Struct2MapWithTrimKV(obj interface{}) (d map[string]interface{}, err error) {
	//先把struct转成json，因为要用json的范式来改变首字母为小写
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		log.GetLogger().Error("Struct2MapWithTrimKV error %v", err)
		return nil, errors.New("Struct2MapWithTrimKV struct to json has error")
	}
	//json转成map
	jsonStr := string(jsonBytes)
	var mapResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &mapResult); err != nil {
		log.GetLogger().Error("Struct2MapWithTrimKV json to map has error %v", err)
		return nil, errors.New("")
	}
	//去掉值为空字符串的key，然后去掉分页字段
	for k, v := range mapResult {
		fmt.Print(v)
		if v == "" {
			delete(mapResult, k)
		}
		if k == "limit" {
			delete(mapResult, k)
		}
		if k == "skip" {
			delete(mapResult, k)
		}
		if k == "pageSize" {
			delete(mapResult, k)
		}
		if k == "page" {
			delete(mapResult, k)
		}
	}
	return mapResult, nil
}

/**
* 结构体转map
* @param   params {[]interface{}} interface数组
* @returns * {[]string} 					string数组
 */
func InterfaceToStringSlice(params []interface{}) []string {
	strArray := make([]string, len(params))
	for i, arg := range params {
		strArray[i] = arg.(string)
	}
	return strArray
}

func GetBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

// 检查文件夹是否存在并新建
func CheckPathExists(folderName string) (string, bool) {
	filePath := config.GetConf("upload.address") + folderName + "/"
	// 路径不存在，创建文件夹
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		log.GetLogger().Error("mkdir has error", err.Error())
		return "", false
	} else {
		log.GetLogger().Infof("mkdir success: %s", filePath)
		return filePath, true
	}
}

func ValidPhone(p string) bool {
	rex := `^(1(([35][0-9])|[8][0-9]|[9][0-9]|[6][0-9]|[7][01356789]|[4][579]))\d{8}$`
	reg := regexp.MustCompile(rex)
	return reg.MatchString(p)
}

func ValidEmail(e string) bool {
	rex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	reg := regexp.MustCompile(rex)
	return reg.MatchString(e)
}

func ValidAccountId(id string) bool {
	rex := `^P\d{12}$`
	reg := regexp.MustCompile(rex)
	return reg.MatchString(id)
}

func ValidAccountStatus(id string) bool {
	if id == "try" || id == "paid" || id == "stop" {
		return true
	}
	return false
}

// ValidPwd 校验密码，8-16个字符，至少包含1个大写字母、1个小写字母、1个数字
func ValidPwd(e string) bool {
	rex := `^(?=.[a-z])(?=.[A-Z])(?=.*\d)[\s\S]{8,16}$`
	reg := regexp.MustCompile(rex)
	return reg.MatchString(e)
}

func ValidTimeFormat(t string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", t)
	if err != nil {
		return false
	}
	return true
}

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
