package main

import (
	"bytes"
	"errors"
	"fmt"
	"geekbang-go/weekThree/gen/annotation"
	"geekbang-go/weekThree/gen/http"

	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// 实际上 main 函数这里要考虑接收参数
// src 源目标
// dst 目标目录
// type src 里面可能有很多类型，那么用户可能需要指定具体的类型
// 这里我们简化操作，只读取当前目录下的数据，并且扫描下面的所有源文件，然后生成代码
// 在当前目录下运行 go install 就将 main 安装成功了，
// 可以在命令行中运行 gen
// 在 testdata 里面运行 gen，则会生成能够通过所有测试的代码

var (
	serviceName = "ServiceName"
	httpClient  = "HttpClient"
	path        = "Path"
)

func main() {
	err := gen(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("success")
}

func gen(src string) error {
	// 第一步找出符合条件的文件
	srcFiles, err := scanFiles(src)
	if err != nil {
		return err
	}
	// 第二步，AST 解析源代码文件，拿到 service definition 定义
	defs, err := parseFiles(srcFiles)
	if err != nil {
		return err
	}
	// 生成代码
	return genFiles(src, defs)
}

// 根据 defs 来生成代码
// src 是源代码所在目录，在测试里面它是 ./testdata
func genFiles(src string, defs []http.ServiceDefinition) error {
	tpl := template.New("service")
	parse, err := tpl.Parse(http.ServiceTpl)
	if err != nil {
		return err
	}
	for _, def := range defs {
		bf := &bytes.Buffer{}
		err := parse.Execute(bf, def)
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		f := abs + string(os.PathSeparator) + "testdata"
		_, err = os.Stat(f)
		if errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(f, 0755)
			if err != nil {
				return err
			}
		}
		filename := f + string(os.PathSeparator) + underscoreName(def.Name) + "_gen.go"
		err = os.WriteFile(filename, bf.Bytes(), 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseFiles(srcFiles []string) ([]http.ServiceDefinition, error) {
	defs := make([]http.ServiceDefinition, 0, 20)
	for _, src := range srcFiles {
		fmt.Println(src)
		// 你需要利用 annotation 里面的东西来扫描 src，然后生成 file
		set := token.NewFileSet()
		f, err := parser.ParseFile(set, src, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		tv := &annotation.SingleFileEntryVisitor{}
		ast.Walk(tv, f)
		var file annotation.File = tv.Get()
		for _, typ := range file.Types {
			_, ok := typ.Annotations.Get(httpClient)
			if !ok {
				continue
			}
			def, err := parseServiceDefinition(file.Node.Name.Name, typ)
			if err != nil {
				return nil, err
			}
			defs = append(defs, def)
		}
	}
	return defs, nil
}

// 你需要利用 typ 来构造一个 http.ServiceDefinition
// 注意你可能需要检测用户的定义是否符合你的预期
func parseServiceDefinition(pkg string, typ annotation.Type) (http.ServiceDefinition, error) {

	serName := ""
	if an, ok := typ.Get(serviceName); ok {
		serName = an.Value
	} else {
		serName = typ.Node.Name.Name
	}
	service := http.ServiceDefinition{
		Package: pkg,
		Name:    serName,
		Methods: make([]http.ServiceMethod, 0, len(typ.Fields)),
	}
	for _, field := range typ.Fields {
		methodPath := ""
		if mePath, ok := field.Get(path); ok {
			methodPath = mePath.Value
		} else {
			methodPath = "/" + field.Node.Names[0].Name
		}
		funcType, ok := field.Node.Type.(*ast.FuncType)
		if !ok {
			return http.ServiceDefinition{}, errors.New("")
		}
		params, ok := parseMethodParamsAndResults(funcType.Params.List)
		if !ok || params[0] != "context.Context" {
			return http.ServiceDefinition{}, errors.New("gen: 方法必须接收两个参数，其中第一个参数是 context.Context，第二个参数请求")
		}
		results, ok := parseMethodParamsAndResults(funcType.Results.List)
		if !ok || results[1] != "error" {
			return http.ServiceDefinition{}, errors.New("gen: 方法必须返回两个参数，其中第一个返回值是响应，第二个返回值是error")
		}
		serviceMethod := http.ServiceMethod{
			Name:         field.Node.Names[0].Name,
			Path:         methodPath,
			ReqTypeName:  params[1],
			RespTypeName: results[0],
		}
		service.Methods = append(service.Methods, serviceMethod)
	}

	return service, nil
}

func parseMethodParamsAndResults(list []*ast.Field) ([]string, bool) {
	if len(list) < 2 {
		return nil, false
	}
	arrStr := make([]string, 0, len(list))
	for _, item := range list {
		switch par := item.Type.(type) {
		case *ast.SelectorExpr:
			pre := par.X.(*ast.Ident).Name
			suf := par.Sel.Name
			arrStr = append(arrStr, pre+"."+suf)
		case *ast.StarExpr:
			pre := par.X.(*ast.Ident).Name
			arrStr = append(arrStr, pre)
		case *ast.Ident:
			arrStr = append(arrStr, par.Name)
		}
	}
	if len(arrStr) < 2 {
		return nil, false
	}
	return arrStr, true
}

// 返回符合条件的 Go 源代码文件，也就是你要用 AST 来分析这些文件的代码
func scanFiles(src string) ([]string, error) {
	arrStr := make([]string, 0)
	err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_service.go") {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			arrStr = append(arrStr, abs)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return arrStr, err
}

// underscoreName 驼峰转字符串命名，在决定生成的文件名的时候需要这个方法
// 可以用正则表达式，然而我写不出来，我是正则渣
func underscoreName(name string) string {
	var buf []byte
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}
