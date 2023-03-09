package controller

import (
	"reflect"
)

var (
	_typeOfError reflect.Type = reflect.TypeOf((*error)(nil)).Elem()
	_nilError                 = reflect.Zero(_typeOfError)
)

func build(target interface{}, kind interface{}) interface{} {

	paramTypes, remap := parameters(target)
	resultTypes, _ := currentResultTypes(target, kind)

	origFn := reflect.ValueOf(target)
	newFnType := reflect.FuncOf(paramTypes, resultTypes, false)
	newFn := reflect.MakeFunc(newFnType, func(args []reflect.Value) []reflect.Value {
		args = remap(args)
		values := origFn.Call(args)

		for _, v := range values {
			values = append(values, v.Convert(reflect.TypeOf(kind).Elem()))
		}

		return values
	})
	return newFn.Interface()
}

func currentResultTypes(target interface{}, kind interface{}) (resultTypes []reflect.Type, hasError bool) {
	ft := reflect.TypeOf(target)
	numOut := ft.NumOut()
	resultTypes = make([]reflect.Type, numOut)

	for i := 0; i < numOut; i++ {
		resultTypes[i] = ft.Out(i)
		if resultTypes[i] == _typeOfError && i == numOut-1 {
			hasError = true
		}
	}
	for i := 0; i < numOut; i++ {
		resultTypes = append(resultTypes, reflect.TypeOf(kind).Elem())
	}
	return resultTypes, hasError
}

func parameters(target interface{}) (
	types []reflect.Type,
	remap func([]reflect.Value) []reflect.Value,
) {
	ft := reflect.TypeOf(target)
	types = make([]reflect.Type, ft.NumIn())
	for i := 0; i < ft.NumIn(); i++ {
		types[i] = ft.In(i)
	}

	return types, func(args []reflect.Value) []reflect.Value {
		return args
	}
}
