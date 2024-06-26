package document

import (
	"context"
	"fmt"
	"monkeylang-server/monkeylang"
	"slices"

	sitter "github.com/smacker/go-tree-sitter"
)

type Document struct {
	Parser                  *sitter.Parser
	Content                 string
	Uri                     string
	byteContent             []byte
	Tree                    *sitter.Tree
	CompletionTree          *sitter.Tree
	NeedsReFreshDiagnostics bool
}

func New(content string) *Document {
	parser := sitter.NewParser()
	parser.SetLanguage(monkeylang.GetLanguage())
	tree, _ := parser.ParseCtx(context.Background(), nil, []byte(content))
	return &Document{
		Parser:         parser,
		Content:        content,
		Uri:            "",
		byteContent:    []byte(content),
		Tree:           tree,
		CompletionTree: nil,
	}
}

func (d *Document) ApplyContentChanges(change string) error {
	d.Content = change
	d.byteContent = []byte(change)

	newTree, err := d.Parser.ParseCtx(context.Background(), nil, d.byteContent)
	if err != nil {
		return fmt.Errorf("could not apply changes=%s", change)
	}

	d.Tree = newTree

	return nil
}

type HighLight struct {
	Line           uint32
	StartChar      uint32
	Length         uint32
	TokenType      string
	TokenModifiers []string
}

func (d *Document) GetHighLights() ([]HighLight, error) {
	highls := make([]HighLight, 0)

	tokens, err := d.queryTokens()
	if err != nil {
		return highls, fmt.Errorf("error getting relevant tokens for=%s", d.Content)
	}

	var prevStart uint32 = 0
	var prevLine uint32 = 0

	for _, tok := range tokens {
		//fmt.Println(tok.Content(d.byteContent))

		t := tok.Type

		l := tok.Start.Row
		if l != prevLine {
			prevStart = 0
		}
		c := tok.Start.Column
		length := tok.End.Column - tok.Start.Column
		ms := []string{}

		deltaChar := c - prevStart
		deltaLine := l - prevLine

		prevStart = c
		prevLine = l

		h := HighLight{
			Line:           deltaLine,
			StartChar:      deltaChar,
			Length:         length,
			TokenType:      t,
			TokenModifiers: ms,
		}
		highls = append(highls, h)
	}

	return highls, nil
}

func (d *Document) GetMethodCompletions(pos DocumentPosition) ([]MethodData, error) {
	ms, err := d.queryAvailableMethods(pos)
	return ms, err
}

type Error struct {
	Start   Position
	End     Position
	Message string
}

func (d *Document) querySyntaxErrors() ([]Error, error) {
	errors := []Error{}
	qs := `
		(ERROR) @error
	`
	q, err := sitter.NewQuery([]byte(qs), monkeylang.GetLanguage())
	if err != nil {
		panic(err)
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, d.Tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range m.Captures {
			e := Error{
				Start: Position{
					Row:    c.Node.StartPoint().Row,
					Column: c.Node.StartPoint().Column,
				},
				End: Position{
					Row:    c.Node.EndPoint().Row,
					Column: c.Node.EndPoint().Column,
				},
			}
			errors = append(errors, e)
		}
	}

	return errors, nil
}

func (d *Document) queryBinaryOpErrors() ([]Error, error) {
	errors := []Error{}
	qs := `
	[
		(binary_expression
			(boolean) 
			(_) 
			[
				(number)
				(string_literal)
			] 
		) 

		(binary_expression
			(string_literal) 
			(_)
			[
				(number)
				(boolean)
			] 
		)

		(binary_expression
			(number) 
			(_)
			[
				(string_literal)
				(boolean)
			] 
		) 
	] @binop

	 
	`
	q, err := sitter.NewQuery([]byte(qs), monkeylang.GetLanguage())
	if err != nil {
		panic(err)
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, d.Tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		m = qc.FilterPredicates(m, d.byteContent)
		for _, c := range m.Captures {
			opNode := c.Node.Child(1)
			if opNode == nil {
				continue
			}
			e := Error{
				Start: Position{
					Row:    opNode.StartPoint().Row,
					Column: opNode.StartPoint().Column,
				},
				End: Position{
					Row:    opNode.EndPoint().Row,
					Column: opNode.EndPoint().Column,
				},
			}
			errors = append(errors, e)
			break
		}
	}

	return errors, nil
}

func (d *Document) queryTokens() ([]SemanticToken, error) {

	tokens := []SemanticToken{}
	qs := `
				"let" @keyword
				"fn" @keyword
				"if" @keyword
				"else" @keyword
				(string_literal) @string_literal
				(let_statement
					left: (identifier) @var_name
					"=" @equals
					right: _ @rest) @let_statement
				(number) @number 
				(boolean) @bool
				(parameter) @parameter
				(function_call
					(identifier) @func_name
					_ @rest) @function_call

	`

	q, err := sitter.NewQuery([]byte(qs), monkeylang.GetLanguage())
	if err != nil {
		panic(err)
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, d.Tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		//m = qc.FilterPredicates(m, []byte(d.Content))
		for _, c := range m.Captures {

			if c.Node.Type() == "let_statement" {
				st := getVarNameToken(c)
				tokens = append(tokens, st)
				break
			}

			if c.Node.Type() == "function_call" {
				st := getFuncNameToken(c)
				tokens = append(tokens, st)
				break
			}

			st := SemanticToken{}
			st.Start = Position{
				c.Node.StartPoint().Row,
				c.Node.StartPoint().Column,
			}
			st.End = Position{
				c.Node.EndPoint().Row,
				c.Node.EndPoint().Column,
			}
			st.Type = c.Node.Type()

			tokens = append(tokens, st)
		}
	}
	qc.Close()

	return tokens, nil

}

type MethodData struct {
	Name string
}

func (d *Document) queryAvailableMethods(triggerPos DocumentPosition) ([]MethodData, error) {
	methods := []MethodData{}
	// TODO query methods that have at least one parameter
	qs := `
				(let_statement
					left: (identifier) @func_name
					"=" 
					right: (function
								(parameters
									(parameter)+ 
								) 
								(body) 
						   ) @function 
				) @let_statement
	`

	q, err := sitter.NewQuery([]byte(qs), monkeylang.GetLanguage())
	if err != nil {
		return methods, err
	}
	qc := sitter.NewQueryCursor()
	qStart := sitter.Point{Row: 1, Column: 1}
	qEnd := sitter.Point{Row: triggerPos.Line, Column: triggerPos.Char}
	qc.SetPointRange(qStart, qEnd)
	qc.Exec(q, d.Tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		//m = qc.FilterPredicates(m, []byte(d.Content))
		for _, c := range m.Captures {

			if c.Node.Type() == "let_statement" {
				md := getFuncName(c, d.byteContent)
				methods = append(methods, md)
				break
			}

		}
	}
	qc.Close()

	return methods, nil
}

func getVarNameToken(c sitter.QueryCapture) SemanticToken {

	st := SemanticToken{}
	identifierNode := c.Node.ChildByFieldName("left")

	st.Start = Position{
		identifierNode.StartPoint().Row,
		identifierNode.StartPoint().Column,
	}

	st.End = Position{
		identifierNode.EndPoint().Row,
		identifierNode.EndPoint().Column,
	}

	st.Type = identifierNode.Type()
	return st
}

func getFuncNameToken(c sitter.QueryCapture) SemanticToken {
	st := SemanticToken{}
	functionNameNode := c.Node.Child(0)
	st.Start = Position{
		functionNameNode.StartPoint().Row,
		functionNameNode.StartPoint().Column,
	}
	st.End = Position{
		functionNameNode.EndPoint().Row,
		functionNameNode.EndPoint().Column,
	}
	st.Type = functionNameNode.Type()
	return st

}

func getFuncName(c sitter.QueryCapture, input []byte) MethodData {
	md := MethodData{}
	n := c.Node.Child(1)
	functionName := n.Content(input)
	md.Name = functionName
	return md

}

type DocumentDiagnosticSeverity int

const (
	ERROR       DocumentDiagnosticSeverity = 1
	WARNING     DocumentDiagnosticSeverity = 2
	INFORMATION DocumentDiagnosticSeverity = 3
	HINT        DocumentDiagnosticSeverity = 4
)

type DocumentPosition struct {
	Line uint32
	Char uint32
}

type Diagnostic struct {
	Start   DocumentPosition
	End     DocumentPosition
	Severty DocumentDiagnosticSeverity
	Source  string
	Message string
}

func (d *Document) GetDiagnostics() []Diagnostic {
	syntaxErrors, _ := d.querySyntaxErrors()
	binOpErrors, _ := d.queryBinaryOpErrors()

	docErrors := slices.Concat(syntaxErrors, binOpErrors)

	diagnostics := []Diagnostic{}

	for _, e := range docErrors {
		diag := Diagnostic{
			Start: DocumentPosition{
				Line: e.Start.Row,
				Char: e.Start.Column,
			},
			End: DocumentPosition{
				Line: e.End.Row,
				Char: e.End.Column,
			},
			Severty: ERROR,
			Source:  d.Uri,
			Message: "Syntax Error",
		}
		diagnostics = append(diagnostics, diag)
	}

	return diagnostics
}
