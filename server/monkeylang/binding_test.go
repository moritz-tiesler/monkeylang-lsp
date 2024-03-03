package monkeylang

import (
	"context"
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestGrammar(t *testing.T) {
	n, err := sitter.ParseCtx(context.Background(), []byte("let func = fn(a) {a;};let func = fn(a) {a;};"), GetLanguage())
	if err != nil {
		t.Errorf("parsing error")
	}
	fmt.Println(n.String())
	expected := "(source_file (function_declaration (function_name) (function (parameters (parameter)) (body (value_name)))) (function_declaration (function_name) (function (parameters (parameter)) (body (value_name)))))"
	if n.String() != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, n.String())
	}

	p := sitter.NewParser()
	p.SetLanguage(GetLanguage())
	tree, err := p.ParseCtx(context.Background(), nil, []byte("let func = fn(a) {a;};let func = fn(a) {a;};"))
	fmt.Println(tree.RootNode().String())
	if err != nil {
		t.Errorf("parsing error")
	}
	if tree.RootNode().String() != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, tree.RootNode().String())
	}
}
