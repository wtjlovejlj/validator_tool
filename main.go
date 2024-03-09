package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"validate/tool"
)

// MyStruct 建表结构体
type MyStruct struct {
	FieldStr1 string `is-sort:"Default" gorm:"column:field_str1"`
	FieldStr2 string `is-sort:"-"`
	FieldStr3 string `is-sort:"-"`
	FieldStr4 string `is-sort:"str4"`
	FieldStr5 string `is-sort:"str5"`
}

type MyStruct1 struct {
	FieldStr1 string `is-sort:"Default" gorm:"column:field_str"`
	FieldStr2 string `is-sort:"-"`
	FieldStr3 string `is-sort:"-"`
	FieldStr4 string `is-sort:"str4"`
	FieldStr5 string `is-sort:"str5"`
}

// TestStruct 入参结构体
type TestStruct struct {
	SortField1 string
	SortField  string `validate:"is-sort"`
}

func main() {
	s := &TestStruct{
		SortField:  "field_str1",
		SortField1: "1111",
	}
	// 验证
	tool.St = MyStruct1{}
	err := tool.Validate.Struct(s)
	if err != nil {
		for _, err1 := range err.(validator.ValidationErrors) {
			// 翻译
			fmt.Println(err1.Translate(tool.Trans))
		}
	} else {
		fmt.Println("验证通过")
	}
}
