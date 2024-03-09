package tool

import (
	"fmt"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"strings"
)

var Validate *validator.Validate
var Trans ut.Translator
var St interface{}

// 获取排序字段 一共定义了三种获取来源
// 1. 结构体字段上定义 is-sort 标签，is-sort 标签值为 Default  通过使用gorm的column标签来获取字段名
// 2. 结构体字段上定义 is-sort 标签 自定义排序字段
// 3. 结构体字段上定义 gorm 标签  值为"-" 根据字段名转换成sql的命名方式
func getSortField(s interface{}) []string {
	sortField := make([]string, 0)
	for i := 0; i < reflect.TypeOf(s).NumField(); i++ {
		if ParseTagSetting(reflect.ValueOf(s).Type().Field(i).Tag.Get("gorm"), ";")["COLUMN"] != "" && reflect.ValueOf(s).Type().Field(i).Tag.Get("is-sort") == "Default" {
			sortField = append(sortField, ParseTagSetting(reflect.ValueOf(s).Type().Field(i).Tag.Get("gorm"), ";")["COLUMN"])
		} else if reflect.ValueOf(s).Type().Field(i).Tag.Get("is-sort") != "" && reflect.ValueOf(s).Type().Field(i).Tag.Get("is-sort") != "-" && reflect.ValueOf(s).Type().Field(i).Tag.Get("is-sort") != "Default" {
			sortField = append(sortField, reflect.ValueOf(s).Type().Field(i).Tag.Get("is-sort"))
		} else if reflect.ValueOf(s).Type().Field(i).Tag.Get("is-sort") == "-" {
			sortField = append(sortField, getDBFieldName(reflect.ValueOf(s).Type().Field(i).Name))
		} else {
			continue
		}
	}
	return sortField
}

// ValidateMyVal 自定义的is-sort验证函数
func ValidateMyVal(fl validator.FieldLevel) bool {
	//获取排序字段
	sortField := getSortField(St)
	if sortField == nil {
		return false
	}
	for _, v := range sortField {
		if fl.Field().String() == v {
			return true
		}
	}
	return false
}

// getDBFieldName 将给定的字符串转换为数据库字段名的格式
func getDBFieldName(name string) string {
	var result strings.Builder

	// 遍历字符串的每个字符
	for i, char := range name {
		// 如果字符为大写字母
		if char >= 'A' && char <= 'Z' {
			// 如果不是第一个字符，添加下划线
			if i > 0 {
				result.WriteByte('_')
			}
			// 将大写字母转为小写字母，并添加到结果中
			result.WriteByte(byte(char + 32))
		} else {
			// 如果字符不是大写字母，直接添加到结果中
			result.WriteByte(byte(char))
		}
	}

	// 返回转换后的数据库字段名
	return result.String()
}

// ParseTagSetting 解析gorm的tag中的column字段 就是将gorm中的字符串用分号分割成切片
// 然后将每一项用冒号分割成键值对
// 最后将键值对放入map中返回
func ParseTagSetting(str string, sep string) map[string]string {
	settings := map[string]string{}
	names := strings.Split(str, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}

func init() {
	Validate = validator.New()
	// 中文翻译器
	uniTrans := ut.New(zh.New())
	Trans, _ = uniTrans.GetTranslator("zh")
	// 注册翻译器到验证器
	err := zhTrans.RegisterDefaultTranslations(Validate, Trans)
	if err != nil {
		panic(fmt.Sprintf("registerDefaultTranslations fail: %s\n", err.Error()))
	}
	// 注册自定义翻译器规则
	_ = Validate.RegisterTranslation("is-sort", Trans, func(ut ut.Translator) error {
		return ut.Add("is-sort", "{0}的值{1}必须是{2}中的一个", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("is-sort", fe.Field(), fmt.Sprintf("%v", fe.Value()), fmt.Sprintf("%s", getSortField(St)))
		return t
	})
	// 注册自定义验证器
	err = Validate.RegisterValidation("is-sort", ValidateMyVal)
	if err != nil {
		fmt.Println(err)
	}
}
