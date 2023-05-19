package tool

import "reflect"

// 通过反射获取对象或指针具体类型
func TypeOf(data interface{}) reflect.Type {
	tof := reflect.TypeOf(data)
	if tof.Kind() == reflect.Ptr {
		tof = tof.Elem()
	}
	return tof
}

// 通过反射获取对象或指针具体类型
func ValueOf(data interface{}) reflect.Value {
	vof := reflect.ValueOf(data)
	if vof.Kind() == reflect.Ptr {
		vof = vof.Elem()
	}
	return vof
}
