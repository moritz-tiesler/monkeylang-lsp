package monkeylang

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
	expected := "(source_file (value_assignment (declaration_name) (number)))"
	treeString := doc.Tree.RootNode().String()
	if treeString != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, treeString)
	}
}

func TestValHighLights(t *testing.T) {
	doc := New("let myVal = true\nlet another = false")

	highlights, err := doc.GetHighLights()
	// fmt.Println(highlights)

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
	//fmt.Println(doc.Tree.RootNode().String())
	//fmt.Println(highlights)

	if err != nil {
		t.Errorf("error getting highlights for %s", doc.Tree.RootNode().String())
	}

	if len(highlights) != 5 {
		t.Errorf("expected 5 highlights from %s, got=%d", doc.Content, len(highlights))
	}
}

func TestQueryTokens(t *testing.T) {
	doc := New("let myVal = 1;let myVal = anotherVal")
	tokens, err := doc.queryTokens()

	//for _, t := range tokens {
	//fmt.Println(t.Content(doc.byteContent))
	//}
	if err != nil {
		t.Errorf("error getting tokens from=%s", doc.Content)
	}
	if len(tokens) != 5 {
		t.Errorf("expected 5 tokens, got=%d", len(tokens))
	}
}

func TestQueryFuncTokens(t *testing.T) {
	doc := New("let func = fn(a) {a;};let func = fn(a) {a;};double(2);")

	fmt.Printf("from parser: \n%s\n", doc.Tree.RootNode().String())
	fmt.Printf("from root node: \n%s\n", doc.Root.String())
	tokens, err := doc.queryTokens()

	if err != nil {
		t.Errorf("error getting tokens from=%s", doc.Content)
	}

	expected_types := []string{
		"let",
		"function_name",
		"fn",
		"parameter",
		"number",
		"let",
		"function_name",
		"fn",
		"parameter",
		"number",
		"function_name",
		"number",
	}

	if len(tokens) != len(expected_types) {
		t.Errorf("expected %d tokens, got=%d", len(expected_types), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type() != expected_types[i] {
			t.Errorf("token type error. want=%s, got=%s", expected_types[i], tok.Type())
		}
	}

	//for _, t := range tokens {
	//	fmt.Println(t.Content(doc.byteContent))
	//}
}
