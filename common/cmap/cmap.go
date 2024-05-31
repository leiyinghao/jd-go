package cmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// type Map map[string]interface{}
type CMap map[string][]string

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
func (v CMap) Get(key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// return Get value and Del key
func (v CMap) Pop(key string) string {
	s := v.Get(key)
	v.Del(key)
	return s
}

// Set sets the key to value. It replaces any existing
// values.
func (v CMap) Set(key, value string) CMap {
	v[key] = []string{value}
	return v
}

// Add adds the value to key. It appends to any existing
// values associated with key.
func (v CMap) Add(key, value string) CMap {
	v[key] = append(v[key], value)
	return v
}

// Del deletes the values associated with key.
func (v CMap) Del(key string) CMap {
	delete(v, key)
	return v
}

// Obtain get all values associated with the given key.
func (v CMap) Obtain(key string) []string {
	if v == nil {
		return []string{}
	}
	return v[key]
}

// Append set the key to value if it doesn't exists. append if it exists.
func (v CMap) Append(key, value string) CMap {
	vs := v.Get(key)
	if vs == "" || len(strings.Trim(vs, " ")) == 0 {
		v.Set(key, value)
		return v
	}
	return v.Set(key, vs+value)
}

// CMap to struct or map[string]string or map[string]interface{} data
func (v CMap) ToStruct(value interface{}) error {
	var m = make(map[string]interface{})
	for k, v := range v {
		m[k] = v[0]
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() == reflect.Struct || t.Kind() == reflect.Map {
			err = json.Unmarshal(b, value)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("Type mismatch. Must be a pointer to Map or structure")
}

func (v CMap) ToValues() url.Values {
	return url.Values(v)
}

func (v CMap) ToMap() map[string]interface{} {
	param := make(map[string]interface{})
	_ = v.ToStruct(&param)
	return param
}

// new a CMap
func New() CMap {
	return CMap{}
}

// new a CMap from url.Values
func Values(values url.Values) CMap {
	return CMap(values)
}

// new a CMap from struct
func Struct(v interface{}) (values CMap) {
	values = New()
	iVal := reflect.ValueOf(v)
	if iVal.Kind() == reflect.Ptr {
		iVal = iVal.Elem()
	}
	// 判断如果传入参数 v 是 map[string]interface{} 调用 paramsToURLValues 否则会 panic
	if iVal.Kind() == reflect.Map && iVal.Type().Key().Kind() == reflect.String && iVal.Type().Elem().Kind() == reflect.Interface {
		// 转换为 map[string]interface{} 类型
		mapVal := iVal.Interface().(map[string]interface{})
		// 调用处理函数 paramsToURLValues 假设这个函数已经定义好
		urlValues, err := paramsToURLValues(mapVal)
		if err != nil {
			fmt.Println("[error] convert params to CMap err ", err.Error())
			return values
		}
		return CMap(urlValues) // 将结果转换并设置到 CMap
	}
	if iVal.Kind() != reflect.Struct {
		fmt.Printf("出错了，传入的参数 v【%v】，类型[%v]不是结构体或者指针\n", v, iVal.Kind().String())
		return
	}
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		fi := typ.Field(i)
		name := fi.Tag.Get("json")
		if name == "" {
			name = fi.Name
		}
		// add support slice
		if iVal.Field(i).Kind() == reflect.Slice {
			var buf bytes.Buffer
			buf.WriteString("[")
			iValArr := iVal.Field(i)
			for j := 0; j < iValArr.Len(); j++ {
				buf.WriteString(fmt.Sprint(`"`, iValArr.Index(j), `",`))
			}
			val := string(buf.Bytes()[:buf.Len()-1])
			val += "]"
			values.Set(name, val)
			continue
		}
		values.Set(name, fmt.Sprint(iVal.Field(i)))
	}
	return
}

// 封装param转为url.Values的函数
func paramsToURLValues(param map[string]interface{}) (url.Values, error) {
	urlParams := url.Values{}
	for key, value := range param {
		switch v := value.(type) {
		case string:
			urlParams.Add(key, v)
		case []string:
			for _, item := range v {
				urlParams.Add(key, item)
			}
		default:
			// 对于非字符串或字符串切片类型的值，需要转换或处理
			// 这里以字符串形式添加，根据实际情况调整
			urlParams.Add(key, fmt.Sprintf("%v", v))
		}
	}

	return urlParams, nil
}
