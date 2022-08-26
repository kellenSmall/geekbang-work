package sql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")

type Insert interface {
	String() string
	Values() []any
	Err() error
	Parse(value reflect.Value)
}

var _ Insert = NewInsertSQL()

func InsertStmt(entity any) (query string, values []any, err error) {
	build := NewInsertSQL()
	build.Parse(reflect.ValueOf(entity))
	return build.String(), build.Values(), build.Err()
}

type insertSQL struct {
	fields []string
	values []any
	name   string
	exist  map[string]struct{}
	err    error
}

func (i *insertSQL) Parse(value reflect.Value) {
	if !value.IsValid() {
		i.err = errInvalidEntity
		return
	}
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() == reflect.Ptr || value.Kind() != reflect.Struct || value.NumField() == 0 {
		i.err = errInvalidEntity
		return
	}

	if i.name == "" {
		i.name = value.Type().Name()
	}

	for j := 0; j < value.NumField(); j++ {
		rfVal := value.Field(j)
		name := value.Type().Field(j).Name
		if i.hasField(name) {
			continue
		}
		anonymous := value.Type().Field(j).Anonymous
		implements := rfVal.Type().Implements(reflect.TypeOf(new(driver.Valuer)).Elem())
		// 是一个struct 是匿名struct 没有实现 某个接口
		if rfVal.Kind() == reflect.Struct && anonymous && !implements {
			i.Parse(rfVal)
			continue
		}
		i.addFieldAndValue(name, rfVal.Interface())
	}

}

func NewInsertSQL() Insert {
	return &insertSQL{exist: map[string]struct{}{}}
}

func (i *insertSQL) hasField(name string) bool {
	_, ok := i.exist[name]
	return ok
}

func (i *insertSQL) String() string {
	if i.name == "" || i.values == nil || i.err != nil {
		return ""
	}
	return fmt.Sprintf(
		"INSERT INTO `%s`(%s) VALUES(%s);",
		i.name,
		strings.Join(i.fields, ","),
		strings.TrimRight(strings.Repeat("?,", len(i.fields)), ","),
	)
}

func (i *insertSQL) Values() []any {
	if i.err != nil {
		return nil
	}
	return i.values
}

func (i *insertSQL) Err() error {
	return i.err
}

func (i *insertSQL) addFieldAndValue(name string, value any) {
	i.fields = append(i.fields, "`"+name+"`")
	i.exist[name] = struct{}{}
	i.values = append(i.values, value)
}
