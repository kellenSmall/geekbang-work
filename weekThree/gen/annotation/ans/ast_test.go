// 看一循环的顺序
package ans

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

type FileVisitors struct {
}

func (f *FileVisitors) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.File:
		fmt.Println("File :", n.Name)
	case *ast.TypeSpec:
		fmt.Println("TypeSpec: ", n.Name)
	case *ast.Field:
		fmt.Printf("Field : %v\n", n.Names)
		return nil
	}
	return f
}

func TestVisit(t *testing.T) {

	set := token.NewFileSet()
	f, err := parser.ParseFile(set, "srv.go", `
// annotation go through the source code and extra the annotation
// @author Deng Ming
/* @multiple first line
second line
*/
// @date 2022/04/02
package annotation

type (
	// FuncType is a type
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/02
	FuncType func()
)

type (
	// StructType is a test struct
	//
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/02
	StructType struct {
		// Public is a field
		// @type string
		Public string
	}

	// SecondType is a test struct
	//
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/03
	SecondType struct {
	}
)

type (
	// Interface is a test interface
	// @author Deng Ming
	/* @multiple first line
	   second line
	*/
	// @date 2022/04/04
	Interface interface {
		// MyFunc is a test func
		// @parameter arg1 int
		// @parameter arg2 int32
		// @return string
		MyFunc(arg1 int, arg2 int32) string

		// second is a test func
		// @return string
		second() string
	}
)
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	ast.Walk(&FileVisitors{}, f)
}
