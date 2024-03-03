package document

import (
	"context"
	"fmt"
	"monkeylang-server/monkeylang"

	sitter "github.com/smacker/go-tree-sitter"
)

type Document struct {
	Parser      *sitter.Parser
	Content     string
	Uri         string
	byteContent []byte
	Tree        *sitter.Tree
}

func New(content string) *Document {
	parser := sitter.NewParser()
	parser.SetLanguage(monkeylang.GetLanguage())
	tree, _ := parser.ParseCtx(context.Background(), nil, []byte(content))
	return &Document{
		Parser:      parser,
		Content:     content,
		Uri:         "",
		byteContent: []byte(content),
		Tree:        tree,
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
	Line           int
	StartChar      int
	Length         int
	TokenType      string
	TokenModifiers []string
}

func (d *Document) GetHighLights() ([]HighLight, error) {
	highls := make([]HighLight, 0)

	tokens, err := d.queryTokens()
	if err != nil {
		return highls, fmt.Errorf("error getting relevant tokens for=%s", d.Content)
	}

	prevStart := 0
	prevLine := 0

	for _, tok := range tokens {

		t := tok.Type()

		l := int(tok.StartPoint().Row)
		if l != prevLine {
			prevStart = 0
		}
		c := int(tok.StartPoint().Column)
		length := int(tok.EndPoint().Column - tok.StartPoint().Column)
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

func (d *Document) queryTokens() ([]*sitter.Node, error) {

	tokens := []*sitter.Node{}
	qs := `
				"let" @keyword
				(declaration_name) @variable
				(number) @number 
				(boolean) @bool
				(function_name) @func_name
				"fn" @keyword
				(parameter) @parameter
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
			tokens = append(tokens, c.Node)
		}
	}

	return tokens, nil

}
