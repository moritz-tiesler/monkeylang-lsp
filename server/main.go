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

	"monkeylang-server/document"

	_ "github.com/tliron/commonlog/simple"
)

const lsName = "Monkeylang"

var version string = "0.0.1"
var handler protocol.Handler
var myServer *server.Server
var doc *document.Document = document.New("")

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
		TextDocumentSemanticTokensFull: highlight,
		// TextDocumentSemanticTokensRange:     highlightRange,
		// TextDocumentSemanticTokensFullDelta: highLightRangeDelta,
	}

	myServer = server.NewServer(&handler, lsName, false)

	myServer.RunStdio()

}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	SetTextDocumentSyncKind(&capabilities, protocol.TextDocumentSyncKindFull)

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
	uri := params.TextDocument.URI
	if doc.Uri == "" {
		doc.Uri = uri
	}
	myServer.Log.Info(fmt.Sprintf("got didChange: %s", uri))

	if doc.Uri != uri {
		return fmt.Errorf(fmt.Sprintf("client changed uri=%s, internally store uri=%s", uri, doc.Uri))
	}

	changes, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole)
	if !ok {
		panic(fmt.Sprintf("could not decode contentChanges=%v", params.ContentChanges...))
	}

	myServer.Log.Info(fmt.Sprintf("got Change: %+v", params.ContentChanges...))
	doc.ApplyContentChanges(changes)

	myServer.Log.Info(fmt.Sprintf("new content=%s", doc.Content))

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

func SetTextDocumentSyncKind(capa *protocol.ServerCapabilities, kind protocol.TextDocumentSyncKind) error {
	options, ok := capa.TextDocumentSync.(*protocol.TextDocumentSyncOptions)
	if !ok {
		return fmt.Errorf("could not set TextDocumentSyncKind")
	}
	options.Change = &kind

	return nil
}

func highlightRange(context *glsp.Context, params *protocol.SemanticTokensRangeParams) (any, error) {

	myServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return nil, nil
}

func highLightRangeDelta(context *glsp.Context, params *protocol.SemanticTokensDeltaParams) (any, error) {
	myServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return nil, nil
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
