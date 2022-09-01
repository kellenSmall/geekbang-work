// 就是把老师的又写了一遍
package ans

import "go/ast"

/*
	    						*ast.File
								↙    ↓   ↘
				       *ast.t	  *ast.t  		*ast.t
				      ↙  ↓  ↘     ↙  ↓  ↘  		↙  ↓  ↘
					*fd  *fd *fd *fd *fd *fd  *fd *fd  *fd
*/

type Field struct {
	Ans Annotations[*ast.Field]
}

type TypeVisitor struct {
	Ans    Annotations[*ast.TypeSpec]
	Fields []Field
}

func (t *TypeVisitor) Visit(node ast.Node) (w ast.Visitor) {
	field, ok := node.(*ast.Field)
	if ok {
		t.Fields = append(t.Fields, Field{Ans: NewAnnotations(field, field.Doc)})
		return nil
	}
	return t
}
func (t *TypeVisitor) Get() Type {
	return Type{
		Ans:    t.Ans,
		Fields: t.Fields,
	}
}

type Type struct {
	Ans    Annotations[*ast.TypeSpec]
	Fields []Field
}

type FileVisitor struct {
	Ans   Annotations[*ast.File]
	Types []*TypeVisitor
}

type File struct {
	Annotations Annotations[*ast.File]
	Types       []Type
}

func (f *FileVisitor) Visit(node ast.Node) (w ast.Visitor) {
	typ, ok := node.(*ast.TypeSpec)
	if ok {
		res := &TypeVisitor{
			Ans:    NewAnnotations(typ, typ.Doc),
			Fields: make([]Field, 0, 0),
		}
		f.Types = append(f.Types, res)
		return res
	}
	return f
}

func (f *FileVisitor) Get() File {

	types := make([]Type, 0, len(f.Types))
	for _, t := range f.Types {
		types = append(types, t.Get())
	}

	return File{
		Annotations: f.Ans,
		Types:       types,
	}
}

type AstManager struct {
	file *FileVisitor
}

func (a *AstManager) Get() File {
	if a.file != nil {
		return a.file.Get()
	}
	return File{}
}
func (a *AstManager) Visit(node ast.Node) (w ast.Visitor) {
	file, ok := node.(*ast.File)
	if ok {
		a.file = &FileVisitor{
			Ans: NewAnnotations(file, file.Doc),
		}
		return a.file
	}
	return a
}
