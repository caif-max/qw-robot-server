package excel

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"qw-robot-server/common/config"
	"qw-robot-server/common/log"
	"qw-robot-server/common/util"
	"reflect"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

func ExportExcel(result interface{}, filed map[string]string, folderName string) (string, error) {
	values := reflect.ValueOf(result)
	types := reflect.TypeOf(result)

	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")

	if err != nil {
		fmt.Printf(err.Error())
		return "", err
	}

	row = sheet.AddRow()

	for i := 0; i < types.Elem().NumField(); i++ {
		cell = row.AddCell()
		cell.Value = filed[types.Elem().Field(i).Name]
	}

	for i := 0; i < values.Len(); i++ {
		row = sheet.AddRow()
		for j := 0; j < values.Index(i).NumField(); j++ {
			cell = row.AddCell()
			if values.Index(i).Field(j).Type().String() == "int" {
				cell.Value = strconv.FormatInt(values.Index(i).Field(j).Int(), 10)
			} else {
				cell.Value = values.Index(i).Field(j).String()
			}
		}
	}
	checkPathExists(folderName)
	fileName := util.UUID()
	fileNamePath := config.GetConf("down.excel.address") + folderName + "/" + fileName + ".xlsx"
	err = file.Save(fileNamePath)

	if err != nil {
		fmt.Printf(err.Error())
		return "", err
	}

	return fileName, nil
}

func ImportExcel(address string, excelFileName string, structType interface{}, limit int) (interface{}, error) {
	var datas []interface{}
	rtype := reflect.TypeOf(structType)

	if strings.LastIndex(excelFileName, "xlsx") > 0 {
		xlFile, err := xlsx.OpenFile(address + excelFileName)
		if err != nil {
			fmt.Println(err)
		}

		for _, sheet := range xlFile.Sheets {
			if limit != -1 && len(sheet.Rows) > limit {
				return nil, errors.New("too mant rows")
			}

			for i := 1; i < len(sheet.Rows); i++ {
				data := reflect.New(rtype).Elem()
				row := sheet.Rows[i]
				for j := 0; j < rtype.NumField(); j++ {
					if j >= len(row.Cells) {
						break
					}
					data.Field(j).SetString(row.Cells[j].String())
				}
				datas = append(datas, data)
			}
		}
	} else if strings.LastIndex(excelFileName, "csv") > 0 {
		file, err := os.Open(excelFileName)
		file.WriteString("\xEF\xBB\xBF")

		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
		defer file.Close()
		reader := csv.NewReader(file)
		count := 1

		for {
			record, err := reader.Read()

			if count == 1 {
				count++
				continue
			} else if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println("Error:", err)
				return nil, err
			}

			data := reflect.New(rtype).Elem()

			for j := 0; j < rtype.NumField(); j++ {
				if j >= len(record) {
					break
				}
				data.Field(j).SetString(record[j])
			}

			datas = append(datas, data)
		}
	} else {
		return nil, errors.New("file type errpr")
	}

	return datas, nil
}

func Init() {
	logPath := config.GetConf("down.excel.address")
	_, err := os.Stat(logPath)
	if err != nil && os.IsNotExist(err) {
		os.MkdirAll(logPath, os.ModePerm)
	}
}

func getFileName(types string) string {
	//return types + "_" + util.GetNowDateTimeFormatCustom("20060102150405") + ".xlsx"
	return util.UUID()
}

func ExportAllExcel(result []map[string]interface{}, filedArray []string, filed map[string]string, folderName string) (string, error) {
	if result == nil || len(result) < 1 {
		return "", errors.New("parameter error")
	}
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error
	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
		return "", err
	}
	row = sheet.AddRow()
	// 添加表头
	for i := 0; i < len(filedArray); i++ {
		cell = row.AddCell()
		if util.Contains(filedArray[i], filed) {
			cell.Value = filed[filedArray[i]]
			fmt.Println("添加表头 ", filed[filedArray[i]])
		} else {
			fmt.Println("对应表头对应数据不存在 ", filedArray[i])
			return "", errors.New("表头对应数据不存在")
		}
	}
	for i := 0; i < len(result); i++ {
		row = sheet.AddRow()
		for k := 0; k < len(filedArray); k++ {
			cell = row.AddCell()
			cellValue := result[i][filedArray[k]]
			if cellValue != nil && util.IsExpectType(cellValue) == "string" {
				cell.Value = result[i][filedArray[k]].(string)
			} else if cellValue != nil && util.IsExpectType(cellValue) == "int" {
				cell.Value = strconv.Itoa(cellValue.(int))
			} else {
				cell.Value = " "
			}
		}
	}

	// 检查路径是否存在并新建
	checkPathExists(folderName)
	fileName := util.UUID()
	fileNamePath := config.GetConf("down.excel.address") + folderName + "/" + fileName + ".xlsx"
	err = file.Save(fileNamePath)

	if err != nil {
		fmt.Printf(err.Error())
		return "", err
	}
	return fileName, nil
}

func checkPathExists(folderName string) {
	filePath := config.GetConf("down.excel.address") + folderName + "/"
	_, err := os.Stat(filePath)
	if err == nil {
		return
	}
	// 路径不存在，创建文件夹
	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		log.GetLogger().Infof("mkdir success: %s", filePath)
	} else {
		log.GetLogger().Infof("mkdir failed: %s", filePath)
	}
}
