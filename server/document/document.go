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
		Parser:  parser,
		Content: content,
		Uri:     "",
		Tree:    tree,
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
	nodes := sitter.NewIterator(d.Tree.RootNode(), sitter.DFSMode)
	highls := make([]HighLight, 0)

	prevStart := 0
	prevLine := 0

	nodes.ForEach(func(n *sitter.Node) error {

		t := n.Type()

		l := int(n.StartPoint().Row)
		if l != prevLine {
			prevStart = 0
		}
		c := int(n.StartPoint().Column)
		length := int(n.EndPoint().Column - n.StartPoint().Column)
		ms := []string{}

		if t == "let_statement" {
			return nil
		}
		if t == "source_file" {
			return nil
		}
		if t == "=" {
			return nil
		}
		if t == "let" {
			return nil
		}

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
		return nil
	})
	return highls, nil
}

// sourceCode := "let a = 2"
// tree, _ := parser.ParseCtx(context.Background(), nil, []byte(sourceCode))

// n := tree.RootNode()
// fmt.Println(n)
