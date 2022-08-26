package ExerclseReflect

import (
	"errors"
	"reflect"
)

var errInvalidEntity = errors.New("invalid entity")

func setint(val *int, newInt int) error {
	rfval := reflect.ValueOf(val)
	rftyp := rfval.Type()
	if rftyp.Kind() != reflect.Ptr && rftyp.Elem().Kind() != reflect.Int {
		return errInvalidEntity
	}
	rfval = rfval.Elem()
	rfval.Set(reflect.ValueOf(newInt))
	return nil
}

func setbool(val *bool, newBool bool) error {
	rfval := reflect.ValueOf(val)
	rftyp := rfval.Type()
	if rftyp.Kind() != reflect.Ptr && rftyp.Elem().Kind() != reflect.Bool {
		return errInvalidEntity
	}
	rfval = rfval.Elem()
	rfval.Set(reflect.ValueOf(newBool))
	return nil
}

func deduplication[T comparable](val []T) ([]T, error) {
	if val == nil {
		return nil, errors.New("非法参数")
	}
	ch := make(map[T]struct{}, len(val))
	res := make([]T, 0, len(val))
	for _, t := range val {
		if _, ok := ch[t]; !ok {
			ch[t] = struct{}{}
			res = append(res, t)
		}
	}
	return res, nil
}
