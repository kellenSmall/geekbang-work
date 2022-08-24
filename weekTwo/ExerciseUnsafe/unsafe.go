package ExerciseUnsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

var errInvalidEntity = errors.New("invalid entity")

type FieldAccessor interface {
	Field(field string) (int, error)
	SetField(field string, val int) error
}

// 使用unsafe 设置值

type FieldMate struct {
	Offset uintptr
}

type UnsafeAccessor struct {
	fields     map[string]FieldMate
	entityAddr unsafe.Pointer
}

func NewUnsafeAccessor(val any) (*UnsafeAccessor, error) {
	if val == nil {
		return nil, errInvalidEntity
	}
	rfVal := reflect.ValueOf(val)
	rfTyp := rfVal.Type()

	if rfTyp.Kind() != reflect.Ptr || rfTyp.Elem().Kind() != reflect.Struct {
		return nil, errInvalidEntity
	}
	elemTyp := rfTyp.Elem()
	numField := elemTyp.NumField()
	fields := make(map[string]FieldMate, numField)
	for i := 0; i < numField; i++ {
		fd := elemTyp.Field(i)
		fields[fd.Name] = FieldMate{Offset: fd.Offset}
	}

	return &UnsafeAccessor{
		fields:     fields,
		entityAddr: rfVal.UnsafePointer(),
	}, nil
}

func (a *UnsafeAccessor) Field(field string) (int, error) {
	fdMate, ok := a.fields[field]
	if !ok {
		return 0, errInvalidEntity
	}
	ptr := unsafe.Pointer(uintptr(a.entityAddr) + fdMate.Offset)
	if ptr == nil {
		return 0, errInvalidEntity
	}
	res := *(*int)(ptr)
	return res, nil
}

func (a *UnsafeAccessor) SetField(field string, val int) error {
	fdMate, ok := a.fields[field]
	if !ok {
		return errInvalidEntity
	}
	ptr := unsafe.Pointer(uintptr(a.entityAddr) + fdMate.Offset)
	if ptr == nil {
		return errInvalidEntity
	}
	*(*int)(ptr) = val
	return nil
}
