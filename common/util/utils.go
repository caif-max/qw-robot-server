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
	"reflect"
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

// 递归查询所有子部门id
func getDeptIds(deptIds *[]string, allDeps []interface{}, deptId string) {
	for i := 0; i < len(allDeps); i++ {
		dept := allDeps[i].(map[string]interface{})
		if dept["pId"] == deptId {
			*deptIds = append(*deptIds, dept["_id"].(string))
			getDeptIds(deptIds, allDeps, dept["_id"].(string))
		}
	}

}

func setUserLevelCondition2Cust(query map[string]interface{}, field string,
	session map[string]interface{}) {
	scope := session["user"].(map[string]interface{})["moduleUsers"].(map[string]interface{})["customer"]
	lowUsers := make([]string, 0)
	managerIds := getSeaIdsByScope(session)

	if scope != nil && scope != "" {
		if IsExpectType(scope) != "string" {
			scope = scope.([]string)
			lowUsers = scope.([]string)
		}
	}

	if scope != "all" {
		if IsMapKey(query, field) { //指定所属范围查询条件
			if Contains(query[field], managerIds) {

			} else {
				if IsExpectType(query[field]) == "string" {
					count := 0
					for i := 0; i < len(lowUsers); i++ {
						uid := lowUsers[i]

						if uid == query[field].(string) {
							count++
							break
						}
					}
					if count == 0 {
						query[field] = "noExistVal"
					}
				} else {
					query[field] = "noExistVal"
				}
			}
		} else {
			lowUsers = append(lowUsers, managerIds...)
			// query[field]=mongo.B{"$in":lowUsers}
			query[field] = map[string]interface{}{
				"$in": lowUsers,
			}
		}
	}
}

func getSeaIdsByScope(session map[string]interface{}) []string {
	scope := session["user"].(map[string]interface{})["moduleUsers"].(map[string]interface{})["customer"]
	if scope == nil {
		scope = make([]string, 0)
	}
	allSeas := session["user"].(map[string]interface{})["allSeas"]
	if allSeas == nil {
		allSeas = make([]interface{}, 0)
	}

	userId := session["user"].(map[string]interface{})["_id"]

	managerIds := make([]string, 0)

	if scope == "all" {
		for i := 0; i < len(allSeas.([]interface{})); i++ {
			pool := allSeas.([]interface{})[i].(map[string]interface{})
			managerIds = append(managerIds, pool["_id"].(string))
		}

	} else {
		for i := 0; i < len(allSeas.([]interface{})); i++ {
			pool := allSeas.([]interface{})[i].(map[string]interface{})
			managers := pool["managers"].(map[string]interface{})
			if managers == nil {
				managers = make(map[string]interface{})
			}
			if len(managers) > 0 && (managers[userId.(string)] == nil || managers[userId.(string)] == "") {
				managerIds = append(managerIds, pool["_id"].(string))
			}
		}
	}
	return managerIds
}

func IsExpectType(i interface{}) string {
	if i == nil {
		return ""
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.String:
		return "string"
	case reflect.Map:
		return "map"
	case reflect.Array:
		return "array"
	case reflect.Struct:
		return "struct"
	case reflect.Slice:
		return "slice"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Interface:
		return "interface"
	case reflect.Bool:
		return "bool"
	default:
		return ""
	}
}

func setUserLevelCondition2Other(query map[string]interface{}, field string, module string,
	session map[string]interface{}, hasCommon bool, menu interface{}) {

	scope := session["user"].(map[string]interface{})["moduleUsers"].(map[string]interface{})[module]

	lowUsers := make([]interface{}, 0)

	if scope != nil && scope != "" {
		if IsExpectType(scope) != "string" {
			scope = scope.([]interface{})
			lowUsers = scope.([]interface{})
		}
	}
	if scope != "all" && (menu != "business" || menu != "custTask" || menu != "custApproval") {
		if query[field] != nil && query[field] != "" {
			if IsExpectType(query[field]) == "string" {
				count := 0
				for i := 0; i < len(lowUsers); i++ {
					uid := lowUsers[i]
					if uid == query[field] {
						count++
						break
					}
				}

				if hasCommon && query[field] == "empty" {
					qfind := []interface{}{"", nil}
					query[field] = map[string]interface{}{
						"$in": qfind,
					}
				} else {
					if count == 0 {
						query[field] = "noExistVa"
					}
				}
			} else {
				query[field] = "noExistVa"
			}
		} else {
			if hasCommon {
				lowUsers = append(lowUsers, "")
				lowUsers = append(lowUsers, nil)
			}
			query[field] = map[string]interface{}{
				"$in": lowUsers,
			}
		}
	} else {
		if query[field] == "empty" {
			qfind := []interface{}{"", nil}
			query[field] = map[string]interface{}{
				"$in": qfind,
			}
		}
	}
}

// 判断obj是否在target中，target支持的类型arrary,slice,map
func Contains(obj interface{}, target interface{}) bool {
	if target == nil {
		return false
	}
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

// 给正则匹配的特殊字符转译
func EscapeText(text string) string {
	reg := regexp.MustCompile(`(\.|\?|\*|\+|\\|\$|\^)`)
	return reg.ReplaceAllString(text, "\\${1}")
}

func IsMapKey(data map[string]interface{}, key string) bool {
	return reflect.ValueOf(data).MapIndex(reflect.ValueOf(key)).IsValid()
}

func IsStructKey(data struct{}, key string) bool {
	return reflect.ValueOf(data).FieldByName(key).IsValid()
}

func GetDbName(session interface{}) (string, bool) {
	if session != nil && IsExpectType(session) == "map" {
		dbname := session.(map[string]interface{})["account"].(map[string]interface{})["dataDB"]
		if dbname != nil && IsExpectType(dbname) == "string" {
			return dbname.(string), false
		}
	}
	return "", false
}

// 获取当前时间 year-month-day hour:min:second
func GetCurrentDate() string {
	t := time.Now()
	return t.Format("2006-01-02 15:04:05")
}

// 获取当前时间 year-month-day hour:min
func GetCurrentDateMin() string {
	t := time.Now()
	return t.Format("2006-01-02")
}

func GetAccount(session interface{}) (string, bool) {

	if session != nil && IsExpectType(session) == "map" {
		account := session.(map[string]interface{})["account"].(map[string]interface{})["account"]
		if account != nil && IsExpectType(account) == "string" {
			return account.(string), false
		}
	}
	return "", true
}

func GetShortTel(tel string) string {
	//mobile := `1[3,4,5,7,8][0-9]{9}`
	reg := regexp.MustCompile(`1[3,4,5,7,8][0-9]{9}`)
	area3 := []string{"010", "020", "021", "023", "024", "025", "027", "028", "029"}
	//_area3 := strings.Join(area3,"-")
	//matchs := mobile.FindAllSubmatch([]byte(tel),-1)
	if !reg.MatchString(tel) {
		if Contains(tel[0:3], area3) {
			return tel[3:]
		} else {
			return tel[4:]
		}
	}
	return tel
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

func BuildUseData(data map[string]interface{}, field []string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	if data == nil || field == nil || len(data) < 1 || len(field) < 1 {
		return nil
	}

	for i := 0; i < len(field); i++ {
		for j := 0; j < len(data); j++ {
			if Contains(field[i], data) {
				resultMap[field[i]] = data[field[i]]
			}
		}
	}
	return resultMap
}

func GetUserId(session interface{}) (string, bool) {
	if session != nil && IsExpectType(session) == "map" {
		id := session.(map[string]interface{})["user"].(map[string]interface{})["_id"]
		if id != nil && IsExpectType(id) == "string" {
			return id.(string), false
		}
	}
	return "", true
}

/**
* 结构体转map
* @param   obj {interface{}} 结构体参数
* @returns d   {map[string]interface{}} 转换后的map
* @returns err {error} 								  错误
 */
func Struct2Map(obj interface{}) (d map[string]interface{}, err error) {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	d = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		d[t.Field(i).Name] = v.Field(i).Interface()
	}
	err = nil
	return
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
