// 就是把老师的又写了一遍
package ans

import (
	"go/ast"
	"strings"
)

// Annotation 存储注解 key -> value
type Annotation struct {
	Key   string
	Value string
}

// Annotations Annotation的集合
type Annotations[N ast.Node] struct {
	Node N
	Ans  []Annotation
}

// NewAnnotations 初始化
func NewAnnotations[N ast.Node](node N, doc *ast.CommentGroup) Annotations[N] {
	if doc == nil || len(doc.List) == 0 {
		return Annotations[N]{Node: node}
	}
	ans := make([]Annotation, 0, len(doc.List))
	for _, item := range doc.List {
		text, ok := excludeComment(item.Text)
		if !ok {
			continue
		}
		if strings.HasPrefix(text, "@") {
			aStr := strings.SplitN(text, " ", 2)
			if s := strings.Trim(aStr[0][1:], " "); len(s) == 0 {
				continue
			}
			key := aStr[0][1:]
			value := ""
			if len(aStr) == 2 {
				value = aStr[1]
			}
			ans = append(ans, Annotation{
				Key:   key,
				Value: value,
			})
		}
	}

	return Annotations[N]{
		Node: node,
		Ans:  ans,
	}
}

// Get 返回Annotation
func (an *Annotations[N]) Get(key string) (Annotation, bool) {
	for _, annotation := range an.Ans {
		if annotation.Key == key {
			return annotation, true
		}
	}
	return Annotation{}, false
}

// excludeComment
func excludeComment(com string) (string, bool) {
	if strings.HasPrefix(com, "// ") {
		return com[3:], true
	} else if strings.HasPrefix(com, "/* ") {
		return com[3 : len(com)-2], true
	}
	return "", false
}
