package document

import (
	"context"
	"fmt"
	"monkeylang-server/monkeylang"

	sitter "github.com/smacker/go-tree-sitter"
	protocol "github.com/tliron/glsp/protocol_3_16"
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
	tree, _ := parser.ParseCtx(context.Background(), nil, []byte(""))
	return &Document{
		Parser:  parser,
		Content: content,
		Uri:     "",
		Tree:    tree,
	}
}

type ContentChanges struct {
	Text string `json:"text"`
}

func (d *Document) ApplyContentChanges(changeData protocol.TextDocumentContentChangeEventWhole) error {
	newContent := changeData.Text
	d.Content = newContent

	newTree, err := d.Parser.ParseCtx(context.Background(), d.Tree, []byte(newContent))
	if err != nil {
		return fmt.Errorf("could not apply changes=%s", newContent)
	}

	d.Tree = newTree

	return nil
}

// sourceCode := "let a = 2"
// tree, _ := parser.ParseCtx(context.Background(), nil, []byte(sourceCode))

// n := tree.RootNode()
// fmt.Println(n)
