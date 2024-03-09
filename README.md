

## 由选择排序字段bug引发的学习探索

问题：我在后端实现字段排序的时候选择前端去传递两个字段一个是要排序的字段另一个是排序的方式，但是没有对要排序的字段进行任何限制

需求：然后需要对排序字段进行校验，去限定哪几个字段可以进行排序

## 想法

1.直接对字段写判断语句（if else  / swich case）

2.validate校验器  这是一个开源库 他通过tag去选择校验逻辑

其中oneof 这个tag 实现了对字段值的限定

### oneof 实现思路

总体来说就是将tag后面自己写死的那些常量和这个字段实际的值（前端传递的值）进行比较

看是否在自己写死的那些常量范围内

我想做的是将写死的常量动态获取到

```
// isOneOf函数用于判断给定的字段值是否在一个参数值列表（自己定义的）中
func isOneOf(fl FieldLevel) bool {
    // 解析参数值列表
    vals := parseOneOfParam2(fl.Param())

    // 获取字段值
    field := fl.Field()

    var v string
    //判断类型
    switch field.Kind() {
    case reflect.String:
        v = field.String()
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        v = strconv.FormatInt(field.Int(), 10)
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        v = strconv.FormatUint(field.Uint(), 10)
    default:
        panic(fmt.Sprintf("Bad field type %T", field.Interface()))
    }

    // 遍历参数值列表，判断是否与字段值相等
    for i := 0; i < len(vals); i++ {
        if vals[i] == v {
            return true
        }
    }

    return false
}

```

![image](https://github.com/wtjlovejlj/validator_tool/assets/89171006/a7fdd048-8ebc-4522-a10d-4d104ea0b30f)


#### 测试小案例

```
package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)
// MyStruct ..
type MyStruct struct {
	Str string `validate:"is-awesome=1 2 3 4 5"`
	Int int    `validate:"oneof=1 2 3"`
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate
var fl validator.FieldLevel

func main() {
	validate = validator.New()
	validate.RegisterValidation("is-awesome", ValidateMyVal)
	s := MyStruct{Str: "awesome", Int: 0}
	err := validate.Struct(s)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

// ValidateMyVal implements validator.Func
func ValidateMyVal(fl validator.FieldLevel) bool {
	// name taking precedence over the fields actual name.
	fmt.Println("FieldName", fl.FieldName())
	// StructFieldName returns the struct field's name
	fmt.Println("StructFieldName", fl.StructFieldName())
	// Param returns param for validation against current field
	fmt.Println("Param", fl.Param())
	// GetTag returns the current validations tag name
	fmt.Println("GetTag", fl.GetTag())

	// 获取字段当前值 fl.Field()
	// 获取tag 对应的参数 fl.Param() ，针对unique_name标签 ，不需要参数
	// 获取字段名称 fl.FieldName()
	return fl.Field().String() == "awesome"
}
```

## 相对动态获取

```
// 三种获取 field
field := reflect.TypeOf(obj).FieldByName("Name")
field := reflect.ValueOf(obj).Type().Field(i)  // i 表示第几个字段
field := reflect.ValueOf(&obj).Elem().Type().Field(i)  // i 表示第几个字段

// 获取 Tag
tag := field.Tag 

// 获取键值对
labelValue := tag.Get("label")
labelValue,ok := tag.Lookup("label")
```

代码思路

利用validator库中的自定义字段验证规则+利用反射获取自定义TAG的值，如果自己没有规定字段的值就获取结构体字段然后转换成下划线的方式

```
package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

// MyStruct ..
type MyStruct struct {
	FieldStr1 string `validate:"" is-sort:"str1"`
	FieldStr2 string `validate:"" is-sort:""`
	FieldStr3 string `validate:"" is-sort:""`
	FieldStr4 string `validate:"" is-sort:"str4"`
	FieldStr5 string `validate:"is-sort" is-sort:"str5"`
	Int       int    `validate:"`
}

func main() {
	Validate := validator.New()
	s := MyStruct{
		FieldStr1: "awesome",
		FieldStr2: "awesome",
		FieldStr3: "awesome",
		FieldStr4: "awesome",
		FieldStr5: "field_str3",
		Int:       0}
	Validate.RegisterValidation("is-sort", ValidateMyVal)
	err := Validate.Struct(s)
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Println("验证通过")
	}
	//fmt.Println(reflect.TypeOf(&MyStruct{}).Elem().NumField())
	//fmt.Println(reflect.ValueOf(&MyStruct{}).Elem().NumField())
}

// ValidateMyVal implements validator.Func
func ValidateMyVal(fl validator.FieldLevel) bool {
	sortField := make([]string, 0)
	for i := 0; i < reflect.TypeOf(MyStruct{}).NumField(); i++ {
		if reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") != "" {
			sortField = append(sortField, reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort"))
		} else {
			sortField = append(sortField, getDBFieldName(reflect.ValueOf(MyStruct{}).Type().Field(i).Name))
		}
	}
	if sortField == nil {
		return false
	}
	fmt.Println(sortField)
	for _, v := range sortField {
		if fl.Field().String() == v {
			return true
		}
	}
	return false
}

// getDBFieldName 将驼峰命名转换为数据库下划线命名
func getDBFieldName(name string) string {
	var result strings.Builder

	for i, char := range name {
		if char >= 'A' && char <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteByte(byte(char + 32)) // 转为小写字母
		} else {
			result.WriteByte(byte(char))
		}
	}

	return result.String()
}


```

执行结果

![image](https://github.com/wtjlovejlj/validator_tool/assets/89171006/5ad14c31-b3d1-4fdf-a3b5-18a66b790bde)


### 优化的最终版本（添加错误中文翻译器、自定义错误、获取gorm中column字段的值）

```
package main

import (
	"fmt"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"strings"
)

var validate *validator.Validate
var trans ut.Translator

// MyStruct ..
type MyStruct struct {
	FieldStr1 string `validate:"" is-sort:"Default" gorm:"column:field_str1"`
	FieldStr2 string `validate:"" is-sort:"-"`
	FieldStr3 string `validate:"" is-sort:"-"`
	FieldStr4 string `validate:"" is-sort:"str4"`
	FieldStr5 string `validate:"is-sort" is-sort:"str5"`
	Int       int    `validate:"oneof=1 2 3"`
}

func validateStruct() {
	s := &MyStruct{
		FieldStr1: "awesome",
		FieldStr2: "awesome",
		FieldStr3: "awesome",
		FieldStr4: "awesome",
		FieldStr5: "int1",
		Int:       0}
	// 注册自定义验证器
	validate.RegisterValidation("is-sort", ValidateMyVal)
	// 验证
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			// 翻译
			fmt.Println(err.Translate(trans))
		}
	} else {
		fmt.Println("验证通过")
	}
}

// 获取排序字段 一共定义了三种获取来源
// 1. 结构体字段上定义 is-sort 标签，is-sort 标签值为 Default  通过使用gorm的column标签来获取字段名
// 2. 结构体字段上定义 is-sort 标签 自定义排序字段
// 3. 结构体字段上定义 gorm 标签  值为"-" 根据字段名转换成sql的命名方式
func getSortField() []string {
	sortField := make([]string, 0)
	for i := 0; i < reflect.TypeOf(MyStruct{}).NumField(); i++ {
		if ParseTagSetting(reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("gorm"), ";")["COLUMN"] != "" && reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") == "Default" {
			sortField = append(sortField, ParseTagSetting(reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("gorm"), ";")["COLUMN"])
		} else if reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") != "" && reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") != "-" && reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") != "Default" {
			sortField = append(sortField, reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort"))
		} else if reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") == "-" {
			sortField = append(sortField, getDBFieldName(reflect.ValueOf(MyStruct{}).Type().Field(i).Name))
		} else {
			continue
		}
	}
	return sortField
}

// ValidateMyVal 自定义的is-sort验证函数
func ValidateMyVal(fl validator.FieldLevel) bool {
	//sortField := make([]string, 0)
	//for i := 0; i < reflect.TypeOf(MyStruct{}).NumField(); i++ {
	//	if reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort") != "" {
	//		sortField = append(sortField, reflect.ValueOf(MyStruct{}).Type().Field(i).Tag.Get("is-sort"))
	//	} else {
	//		sortField = append(sortField, getDBFieldName(reflect.ValueOf(MyStruct{}).Type().Field(i).Name))
	//	}
	//}
	//获取排序字段
	sortField := getSortField()
	if sortField == nil {
		return false
	}
	fmt.Println(sortField)
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

func main() {
	validate = validator.New()

	// 中文翻译器
	uniTrans := ut.New(zh.New())
	trans, _ = uniTrans.GetTranslator("zh")
	// 注册翻译器到验证器
	err := zhTrans.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic(fmt.Sprintf("registerDefaultTranslations fail: %s\n", err.Error()))
	}
	// 注册自定义翻译器规则
	_ = validate.RegisterTranslation("is-sort", trans, func(ut ut.Translator) error {
		return ut.Add("is-sort", "{0}的值{1}必须是{2}中的一个", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("is-sort", fe.Field(), fmt.Sprintf("%v", fe.Value()), fmt.Sprintf("%s", getSortField()))
		return t
	})
	validateStruct()
	//fmt.Println(reflect.TypeOf(&MyStruct{}).Elem().NumField())
	//fmt.Println(reflect.ValueOf(&MyStruct{}).Elem().NumField())
	//fmt.Println(reflect.TypeOf(&MyStruct{}))
	//fmt.Println(reflect.ValueOf(&MyStruct{}))
	//fmt.Println(reflect.ValueOf(&MyStruct{}).Interface())
	//fmt.Println(reflect.ValueOf(&MyStruct{}).Elem().CanSet())
	//fmt.Println(reflect.ValueOf(MyStruct{}).Type().Field(2).Tag.Get("is-sort"))
	//fmt.Println(reflect.ValueOf(MyStruct{}).Type().Field(0).Tag.Get("gorm"))
	//获取gorm中column的值
	//fmt.Println(reflect.ValueOf(MyStruct{}).Type().Field(0).Tag.Get("gorm"))
	//tagSetting := ParseTagSetting(reflect.ValueOf(MyStruct{}).Type().Field(0).Tag.Get("gorm"), ";")
	//fmt.Println(tagSetting["COLUMN"])
}
```



## 反射

https://juejin.cn/post/7183132625580605498#heading-10

大佬笔记

## 反射定律

1. 反射可以将 `interface` 类型变量转换成反射对象。（   reflect.TypeOf()/reflect.ValueOf()    ）
2. 反射可以将反射对象还原成 `interface` 对象。(reflect.ValueOf().Interface())
3. 如果要修改反射对象，那么反射对象必须是可设置的（`CanSet`）。(reflect.ValueOf().Elem().CanSet())

### Elem方法

```
fmt.Println(reflect.TypeOf(&MyStruct{}).Elem())去掉&报panic

fmt.Println(reflect.ValueOf(&MyStruct{}).Elem().CanSet()) 不加Elem方法就不可以修改
```

那什么情况下一个反射对象是可设置的呢？前提是这个反射对象是一个指针，然后这个指针指向的是一个可设置的变量。 在我们传递一个值给 `reflect.ValueOf` 的时候，如果这个值只是一个普通的变量，那么 `reflect.ValueOf` 会返回一个不可设置的反射对象。 因为这个值实际上被拷贝了一份，我们如果通过反射修改这个值，那么实际上是修改的这个拷贝的值，而不是原来的值。 所以 go 语言在这里做了一个限制，如果我们传递进 `reflect.ValueOf` 的变量是一个普通的变量，那么在我们设置反射对象的值的时候，会报错。 所以在上面这个例子中，我们传递了 `x` 的指针变量作为参数。这样，运行时就可以找到 `x` 本身，而不是 `x` 的拷贝，所以就可以修改 `x` 的值了。

1.反射对象是一个指针

2.传递的时候传指针变量

![reflect_3.png](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/ff7cdc36fdbe4188a630b816b49ca3ca~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp?)

## reflect.Type

![reflect_6.png](https://p9-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/778a012296344e3299d50c5c4c49e722~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp?)

## reflect.Value

![reflect_7.png](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/fce16bde5154483ba644c875be2c8346~tplv-k3u1fbpfcp-zoom-in-crop-mark:1512:0:0:0.awebp?)
