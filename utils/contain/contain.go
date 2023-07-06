package contain

import (
	"errors"
	"reflect"
)

func Contain(list interface{}, obj interface{}) (bool, error) {
	if list == nil {
		return false, nil
	}
	targetValue := reflect.ValueOf(list)
	switch reflect.TypeOf(list).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
		return false, nil
	}
	return false, errors.New("not in array")
}
