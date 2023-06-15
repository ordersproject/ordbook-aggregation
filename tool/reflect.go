package tool

import "reflect"

func TypeOf(data interface{}) reflect.Type {
	tof := reflect.TypeOf(data)
	if tof.Kind() == reflect.Ptr {
		tof = tof.Elem()
	}
	return tof
}

func ValueOf(data interface{}) reflect.Value {
	vof := reflect.ValueOf(data)
	if vof.Kind() == reflect.Ptr {
		vof = vof.Elem()
	}
	return vof
}
