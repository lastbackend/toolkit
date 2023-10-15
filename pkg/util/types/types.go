package types

import "reflect"

func Type(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().PkgPath() + "." + t.Elem().Name()
	} else {
		return t.PkgPath() + "." + t.Name()
	}
}
