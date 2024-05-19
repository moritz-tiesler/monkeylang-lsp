package document

import (
	"testing"
)

func TestApplyChanges(t *testing.T) {
	doc := New("")
	sourceCode := "let a = 2"

	err := doc.ApplyContentChanges(sourceCode)
	if err != nil {
		t.Errorf("could not apply changes=%s", sourceCode)
	}
	expected := "(source_file (let_statement left: (identifier) right: (number)))"
	treeString := doc.Tree.RootNode().String()
	if treeString != expected {
		t.Errorf("grammar error. want=%s, got=%s", expected, treeString)
	}
}

func TestValHighLights(t *testing.T) {
	doc := New("let myVal = true\nlet another = false")

	highlights, err := doc.GetHighLights()
	//fmt.Println(highlights)

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

func TestStringHighlights(t *testing.T) {
	doc := New(`let myStr = "abc"`)

	highlights, err := doc.GetHighLights()
	//fmt.Println(doc.Tree.RootNode().String())
	//fmt.Println(highlights)

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

func TestQueryOperatorErrors(t *testing.T) {
	doc := New(
		`1 + true
"abc" + 1
true + 4
myFunc(1 + true)
myVar + true
`)

	tokens, err := doc.queryBinaryOpErrors()
	if err != nil {
		t.Errorf("error getting tokens from=%s", doc.Content)
	}

	expected := []Error{
		{
			Start: Position{0, 2},
			End:   Position{0, 3},
		},
		{
			Start: Position{1, 6},
			End:   Position{1, 7},
		},
		{
			Start: Position{2, 5},
			End:   Position{2, 6},
		},
		{
			Start: Position{3, 9},
			End:   Position{3, 10},
		},
	}

	if len(tokens) != len(expected) {
		t.Errorf("wrong number of errors: expected=%d, got=%d", len(expected), len(tokens))
	}

	for i, e := range tokens {
		if e != expected[i] {
			t.Errorf("wrong error node: expected=%v, got=%v", expected[i], e)
		}
	}
}

func TestQueryFuncTokens(t *testing.T) {
	doc := New(
		`let double = fn(a) {a * 2;}
		let double = fn(a) {a * 2;}
		double(2)`)
	tokens, err := doc.queryTokens()

	if err != nil {
		t.Errorf("error getting tokens from=%s", doc.Content)
	}

	expected_types := []string{
		"let",
		"identifier",
		"fn",
		"parameter",
		"number",
		"let",
		"identifier",
		"fn",
		"parameter",
		"number",
		"identifier",
		"number",
	}

	if len(tokens) != len(expected_types) {
		t.Errorf("expected %d tokens, got=%d", len(expected_types), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected_types[i] {
			t.Errorf("token type error. want=%s, got=%s", expected_types[i], tok.Type)
		}
	}

	//for _, t := range tokens {
	//fmt.Println(t.Content(doc.byteContent))
	//}
}

func TestQueryStringTokens(t *testing.T) {
	doc := New(
		`let myStr = "abc"`)
	tokens, err := doc.queryTokens()

	if err != nil {
		t.Errorf("error getting tokens from=%s", doc.Content)
	}

	expected_types := []string{
		"let",
		"identifier",
		"string_literal",
	}

	if len(tokens) != len(expected_types) {
		t.Errorf("expected %d tokens, got=%d", len(expected_types), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected_types[i] {
			t.Errorf("token type error. want=%s, got=%s", expected_types[i], tok.Type)
		}
	}

	//for _, t := range tokens {
	//fmt.Println(t.Content(doc.byteContent))
	//}
}

func TestQueryErrors(t *testing.T) {
	doc := New(
		`let myStr == "abc"`)

	tokens, err := doc.querySyntaxErrors()

	if err != nil {
		t.Errorf("error getting errots from=%s", doc.Content)
	}

	expected := []string{
		"ERROR",
	}

	if len(tokens) != len(expected) {
		t.Errorf("expected %d tokens from %s, got=%d",
			len(tokens), doc.Tree.RootNode().String(), len(expected))
	}
}

func TestGetDiagnostics(t *testing.T) {
	doc := New(
		`let myStr == "abc"`)

	diags := doc.GetDiagnostics()

	expected := []Diagnostic{
		{
			Start:   DocumentPosition{0, 11},
			End:     DocumentPosition{0, 12},
			Severty: 1,
			Message: "Syntax Error",
		},
	}

	if len(diags) != len(expected) {
		t.Errorf("wrong number of diagnostics. expected=%d, got=%d", len(expected), len(diags))

	}

	for i, d := range diags {
		if expected[i] != d {
			t.Errorf("wrong diagnostic: expected=%v, got=%v", expected[i], d)
		}
	}
}

func TestQueryAvailableMethods(t *testing.T) {
	doc := New(
		`
let proc = fn() {return 2}
let funcA = fn(a) {a + 2}
let funcB = fn(a, b) { a + b }
let res = 2.
let unavailableFunc = fn(b) {b * 2}
`)

	triggerPos := DocumentPosition{Line: 5, Char: 12}
	methodNames, _ := doc.GetMethodCompletions(triggerPos)

	expected := []MethodData{
		{
			Name: "funcA",
		},
		{
			Name: "funcB",
		},
	}

	if len(methodNames) != len(expected) {
		t.Errorf("wrong number of methods. expected=%d, got=%d", len(expected), len(methodNames))

	}

	for i, d := range methodNames {
		if expected[i] != d {
			t.Errorf("wrong methodname: expected=%v, got=%v", expected[i], d)
		}
	}
}
