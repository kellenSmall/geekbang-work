package annotation

import (
	"go/ast"
)

// SingleFileEntryVisitor 这部分和课堂演示差不多，但是我建议你们自己试着写一些
type SingleFileEntryVisitor struct {
	file *fileVisitor
}

func (s *SingleFileEntryVisitor) Get() File {
	if s.file != nil {
		return s.file.Get()
	}
	return File{}
}

func (s *SingleFileEntryVisitor) Visit(node ast.Node) ast.Visitor {
	file, ok := node.(*ast.File)
	if ok {
		s.file = &fileVisitor{
			ans: newAnnotations[*ast.File](file, file.Doc),
		}
		return s.file
	}
	return s
}

type fileVisitor struct {
	ans     Annotations[*ast.File]
	types   []*typeVisitor
	visited bool
}

func (f *fileVisitor) Get() File {

	typ := make([]Type, 0, len(f.types))
	for _, visitor := range f.types {
		typ = append(typ, visitor.Get())
	}

	return File{
		Annotations: f.ans,
		Types:       typ,
	}
}

func (f *fileVisitor) Visit(node ast.Node) ast.Visitor {
	spec, ok := node.(*ast.TypeSpec)
	if ok {
		res := &typeVisitor{
			ans:    newAnnotations[*ast.TypeSpec](spec, spec.Doc),
			fields: make([]Field, 0, 0),
		}
		f.types = append(f.types, res)
		return res

	}
	return f
}

type File struct {
	Annotations[*ast.File]
	Types []Type
}

type typeVisitor struct {
	ans    Annotations[*ast.TypeSpec]
	fields []Field
}

func (t *typeVisitor) Get() Type {
	return Type{
		Annotations: t.ans,
		Fields:      t.fields,
	}
}

func (t *typeVisitor) Visit(node ast.Node) (w ast.Visitor) {
	field, ok := node.(*ast.Field)
	if ok {
		res := Field{
			Annotations: newAnnotations(field, field.Doc),
		}
		t.fields = append(t.fields, res)
		return nil
	}
	return t
}

type Type struct {
	Annotations[*ast.TypeSpec]
	Fields []Field
}

type Field struct {
	Annotations[*ast.Field]
}
