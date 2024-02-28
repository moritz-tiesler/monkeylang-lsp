package main

import (
	"encoding/json"
	"fmt"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog

	_ "github.com/tliron/commonlog/simple"
)

const lsName = "Monkeylang"

var version string = "0.0.1"
var handler protocol.Handler
var myServer *server.Server

func main() {
	path := "/tmp/lsp.log"
	commonlog.Configure(1, &path)

	handler = protocol.Handler{
		Initialize:            initialize,
		Initialized:           initialized,
		Shutdown:              shutdown,
		SetTrace:              setTrace,
		TextDocumentDidChange: didChange,
		//TextDocumentCompletion:              complete,
		TextDocumentSemanticTokensFull:      highlight,
		TextDocumentSemanticTokensRange:     highlightRange,
		TextDocumentSemanticTokensFullDelta: highLightRangeDelta,
	}

	myServer = server.NewServer(&handler, lsName, false)

	myServer.RunStdio()

	// parser := sitter.NewParser()
	// parser.SetLanguage(monkeylang.GetLanguage())

	// sourceCode := "let a = 2"
	// tree, _ := parser.ParseCtx(context.Background(), nil, []byte(sourceCode))

	// n := tree.RootNode()
	// fmt.Println(n)
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	AddTokenLegend(&capabilities)

	jsonBytes, err := json.Marshal(capabilities)
	if err != nil {
		myServer.Log.Error("error reading capas")
	}
	myServer.Log.Info(fmt.Sprintf("Capas: %+v", string(jsonBytes)))

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

func didChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	myServer.Log.Info(fmt.Sprintf("got smth: %d", params.TextDocument.Version))
	return nil
}

func complete(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	return protocol.CompletionList{
		IsIncomplete: false,
		Items: []protocol.CompletionItem{
			{
				Label: "Monkeylang",
			},
			{
				Label: "LSP",
			},
			{
				Label: "Lua",
			},
		},
	}, nil
}

func highlight(context *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {

	myServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return &protocol.SemanticTokens{
		Data: []uint32{},
	}, nil
}

func AddTokenLegend(h *protocol.ServerCapabilities) {
	h.SemanticTokensProvider.(*protocol.SemanticTokensOptions).Legend = protocol.SemanticTokensLegend{
		TokenTypes: []string{
			string(protocol.SemanticTokenTypeVariable),
			string(protocol.SemanticTokenTypeNamespace),
			string(protocol.SemanticTokenTypeType),
			string(protocol.SemanticTokenTypeClass),
			string(protocol.SemanticTokenTypeEnum),
			string(protocol.SemanticTokenTypeInterface),
			string(protocol.SemanticTokenTypeStruct),
			string(protocol.SemanticTokenTypeTypeParameter),
			string(protocol.SemanticTokenTypeParameter),
			string(protocol.SemanticTokenTypeVariable),
			string(protocol.SemanticTokenTypeProperty),
			string(protocol.SemanticTokenTypeEnumMember),
			string(protocol.SemanticTokenTypeEvent),
			string(protocol.SemanticTokenTypeFunction),
			string(protocol.SemanticTokenTypeMethod),
			string(protocol.SemanticTokenTypeMacro),
			string(protocol.SemanticTokenTypeKeyword),
			string(protocol.SemanticTokenTypeModifier),
			string(protocol.SemanticTokenTypeComment),
			string(protocol.SemanticTokenTypeString),
			string(protocol.SemanticTokenTypeNumber),
			string(protocol.SemanticTokenTypeRegexp),
			string(protocol.SemanticTokenTypeOperator),
		},
		TokenModifiers: []string{
			string(protocol.SemanticTokenModifierDeclaration),
			string(protocol.SemanticTokenModifierDefinition),
			string(protocol.SemanticTokenModifierReadonly),
			string(protocol.SemanticTokenModifierStatic),
			string(protocol.SemanticTokenModifierDeprecated),
			string(protocol.SemanticTokenModifierAbstract),
			string(protocol.SemanticTokenModifierAsync),
			string(protocol.SemanticTokenModifierModification),
			string(protocol.SemanticTokenModifierDocumentation),
			string(protocol.SemanticTokenModifierDefaultLibrary),
		},
	}
}

func highlightRange(context *glsp.Context, params *protocol.SemanticTokensRangeParams) (any, error) {

	myServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return nil, nil
}

func highLightRangeDelta(context *glsp.Context, params *protocol.SemanticTokensDeltaParams) (any, error) {
	myServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return nil, nil
}
