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
	expected := "(source_file (value_assignment (value_name) (number)))"
	treeString := doc.Tree.RootNode().String()
	if treeString != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, treeString)
	}
}

func TestValHighLights(t *testing.T) {
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

func TestFuncHighlights(t *testing.T) {
	doc := New("let double = fn(a) {a*2;};")

	highlights, err := doc.GetHighLights()
	fmt.Println(doc.Tree.RootNode().String())
	fmt.Println(highlights)

	if err != nil {
		t.Errorf("error getting highlights for %s", doc.Tree.RootNode().String())
	}

	if len(highlights) != 3 {
		t.Errorf("expected 3 highlights from %s, got=%d", doc.Content, len(highlights))
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
