package document

import (
	"fmt"
	"testing"
)

func TestApplyChanges(t *testing.T) {
	doc := New("")
	sourceCode := "let a = 2"

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
	if len(highlights) < 1 {
		t.Errorf("expected at least one node from %s", doc.Tree.RootNode().String())
	}
}
