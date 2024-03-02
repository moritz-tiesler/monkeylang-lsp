package document

import (
	"fmt"
	"testing"
)

func TestApplyChanges(t *testing.T) {
	doc := New("")
	sourceCode := "let a = 2;"

	err := doc.ApplyContentChanges(sourceCode)
	if err != nil {
		t.Errorf("could not apply changes=%s", sourceCode)
	}
	expected := "(source_file (let_statement (identifier) (number)))"
	treeString := doc.Tree.RootNode().String()
	if treeString != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, treeString)
	}
}

func TestHighLights(t *testing.T) {
	doc := New("let myVal = true\nlet another = false")
	highlights, err := doc.GetHighLights()
	fmt.Println(highlights)
	if err != nil {
		t.Errorf("error getting highlights for %s", doc.Tree.RootNode().String())
	}
	if len(highlights) != 6 {
		t.Errorf("expected six four nodes from %s", doc.Content)
	}
}

func TestQueryTokens(t *testing.T) {
	doc := New("let myVal = 1;let myVal = anotherVal")
	tokens, err := doc.queryTokens()

	for _, t := range tokens {
		fmt.Println(t.Content(doc.byteContent))
	}
	if err != nil {
		t.Errorf("error getting tokens from=%s", doc.Content)
	}
	if len(tokens) != 6 {
		t.Errorf("expected 6 tokens, got=%d", len(tokens))
	}
}
