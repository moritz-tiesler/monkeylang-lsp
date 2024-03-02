package monkeylang

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestGrammar(t *testing.T) {
	n, err := sitter.ParseCtx(context.Background(), []byte("let a = 2;"), GetLanguage())
	if err != nil {
		t.Errorf("parsing error")
	}
	expected := "(source_file (let_statement (identifier) (number)))"
	if n.String() != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, n.String())
	}

	p := sitter.NewParser()
	p.SetLanguage(GetLanguage())
	node, err := p.ParseCtx(context.Background(), nil, []byte("let a = 2"))
	if err != nil {
		t.Errorf("parsing error")
	}
	if n.String() != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, node.RootNode().String())
	}

}
