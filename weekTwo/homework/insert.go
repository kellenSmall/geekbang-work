package homework

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")

//func InsertStmt(entity interface{}) (string, []interface{}, error) {
//	if entity == nil {
//		return "", nil, errInvalidEntity
//	}
//	/* 代码有点丑
//	val := reflect.ValueOf(entity)
//	typ := reflect.TypeOf(entity)
//	if typ.Kind() != reflect.Struct && (typ.Kind() == reflect.Ptr && typ.Elem().Kind() != reflect.Struct) {
//		return "", nil, errInvalidEntity
//	}
//	if typ.Kind() == reflect.Ptr {
//		typ = typ.Elem()
//		val = val.Elem()
//	}
//	if typ.Kind() == reflect.Struct && typ.NumField() == 0 {
//		return "", nil, errInvalidEntity
//	}
//	// 检测 entity 是否符合我们的要求
//	// 我们只支持有限的几种输入
//	sb := strings.Builder{}
//	sb.WriteString("INSERT INTO ")
//	if typ.Kind() == reflect.Struct {
//		name := typ.Name()
//		fnum := typ.NumField()
//		args := make([]any, 0, fnum)
//		str := ""
//		ps := ""
//		for i := 0; i < fnum; i++ {
//			fd := typ.Field(i)
//			fdVal := val.Field(i)
//			fdt := fdVal.Type()
//			if fdt.Kind() == reflect.Struct && fdt.String() != "sql.NullString" {
//				subnum := fdt.NumField()
//				for j := 0; j < subnum; j++ {
//					field := fdt.Field(j)
//					subfdval := fdVal.Field(j)
//					args = append(args, subfdval.Interface())
//					str += "?,"
//					ps += fmt.Sprintf("`%s`,", field.Name)
//				}
//			}
//			if fdt.Kind() != reflect.Struct || fdt.String() == "sql.NullString" {
//				args = append(args, fdVal.Interface())
//				str += "?,"
//				ps += fmt.Sprintf("`%s`,", fd.Name)
//			}
//		}
//		str = strings.TrimRight(str, ",")
//		ps = strings.TrimRight(ps, ",")
//		sb.WriteString(fmt.Sprintf("`%s`(%s) VALUES(%s);", name, ps, str))
//		return sb.String(), args, nil
//	}*/
//
//	// 使用 strings.Builder 来拼接 字符串
//	// bd := strings.Builder{}
//
//	// 构造 INSERT INTO XXX，XXX 是你的表名，这里我们直接用结构体名字
//
//	// 遍历所有的字段，构造出来的是 INSERT INTO XXX(col1, col2, col3)
//	// 在这个遍历的过程中，你就可以把参数构造出来
//	// 如果你打算支持组合，那么这里你要深入解析每一个组合的结构体
//	// 并且层层深入进去
//
//	// 拼接 VALUES，达成 INSERT INTO XXX(col1, col2, col3) VALUES
//
//	// 再一次遍历所有的字段，要拼接成 INSERT INTO XXX(col1, col2, col3) VALUES(?,?,?)
//	// 注意，在第一次遍历的时候我们就已经拿到了参数的值，所以这里就是简单拼接 ?,?,?
//
//	// return bd.String(), args, nil
//	panic("implement me")
//}

func InsertStmt(entity interface{}) (string, []interface{}, error) {
	if entity == nil {
		return "", nil, errInvalidEntity
	}
	ac := NewAccessor(entity)
	if !ac.HasStruct() && !ac.HasPtrNextStruct() {
		return "", nil, errInvalidEntity
	}
	if !ac.HasStruct() {
		ac.PtrToStruct()
	}
	if ac.HasFields() {
		return "", nil, errInvalidEntity
	}
	str, args := ac.setTableName().rangeFields().builder()
	return str, args, nil
}

type Accessor struct {
	typ         reflect.Type
	val         reflect.Value
	name        string
	fields      []string
	args        []any
	pd          []string
	checkString map[string]struct{}
}

// 代码实现还是有问题
func (ac *Accessor) setTableName() *Accessor {
	ac.name = ac.typ.Name()
	return ac
}
func NewAccessor(val any) *Accessor {
	return &Accessor{
		typ:         reflect.TypeOf(val),
		val:         reflect.ValueOf(val),
		fields:      make([]string, 0),
		pd:          make([]string, 0),
		args:        make([]any, 0),
		checkString: make(map[string]struct{}),
	}
}

// HasPtrNextStruct 当前val 是否是 指针且elem 元素类型是 struct ?
func (ac *Accessor) HasPtrNextStruct() bool {
	return ac.typ.Kind() == reflect.Ptr && ac.typ.Elem().Kind() == reflect.Struct
}

// HasStruct 当前 val 是否是 struct ?
func (ac *Accessor) HasStruct() bool {
	return ac.typ.Kind() == reflect.Struct
}

// HasFields 当前 val 是否有 fields ?
func (ac *Accessor) HasFields() bool {
	return ac.typ.NumField() == 0
}

// PtrToStruct ptr 转 struct
func (ac *Accessor) PtrToStruct() {
	ac.typ = ac.typ.Elem()
	ac.val = ac.val.Elem()
}

// checkField 检查字段是否存在
func (ac *Accessor) checkField(field string) bool {
	_, ok := ac.checkString[field]
	return ok
}

func (ac *Accessor) addField(field string) {
	ac.checkString[field] = struct{}{}

}

// rangeFields 遍历
func (ac *Accessor) rangeFields() *Accessor {
	ac.deep(ac.val)
	return ac
}

// 生成 sql
func (ac *Accessor) builder() (string, []any) {
	//INSERT INTO `Customer`(`CreateTime`,`UpdateTime`,`Id`,`NickName`,`Age`,`Address`,`Company`) VALUES(?,?,?,?,?,?,?);
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("INSERT INTO `%s`", ac.name))
	fieldsStr := strings.Join(ac.fields, ",")
	pds := strings.Join(ac.pd, ",")
	sb.WriteString(fmt.Sprintf("(%s) VALUES(%s);", fieldsStr, pds))
	return sb.String(), ac.args
}

// addData
func (ac *Accessor) addData(field reflect.StructField, value reflect.Value) {
	name := fmt.Sprintf("`%s`", field.Name)
	if !ac.checkField(name) {
		ac.addField(name)
		ac.fields = append(ac.fields, name)
		ac.args = append(ac.args, value.Interface())
		ac.pd = append(ac.pd, "?")
	}
}

func (ac *Accessor) deep(value reflect.Value) {
	ft := value.Type()
	numf := ft.NumField()
	for i := 0; i < numf; i++ {
		fdt := ft.Field(i)
		fdVal := value.Field(i)
		switch fdVal.Type().Kind() {
		case reflect.Struct:
			switch fdVal.Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
			case true:
				ac.addData(fdt, fdVal)
			default:
				if !fdt.Anonymous {
					ac.addData(fdt, fdVal)
				} else {
					ac.deep(fdVal)
				}
			}
		default:
			ac.addData(fdt, fdVal)
		}
	}
}
