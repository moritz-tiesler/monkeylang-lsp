package main

import (
	"encoding/json"
	"fmt"
	"time"

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

type Server struct {
	glspServer *server.Server
}

var myServer Server
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
		TextDocumentCompletion: complete,
	}

	myServer = Server{
		glspServer: server.NewServer(&handler, lsName, false),
	}
	myServer.glspServer.RunStdio()

}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	SetTextDocumentSyncKind(&capabilities, protocol.TextDocumentSyncKindFull)
	AddTokenLegend(&capabilities)

	jsonBytes, err := json.Marshal(capabilities)
	if err != nil {
		myServer.glspServer.Log.Error("error reading capas")
	}
	myServer.glspServer.Log.Info(fmt.Sprintf("Capas: %+v", string(jsonBytes)))

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
	myServer.glspServer.Log.Info(fmt.Sprintf("got didChange: %s", uri))

	if doc.Uri != uri {
		return fmt.Errorf(fmt.Sprintf("client changed uri=%s, internally store uri=%s", uri, doc.Uri))
	}

	changes, ok := params.ContentChanges[0].(protocol.TextDocumentContentChangeEventWhole)
	if !ok {
		panic(fmt.Sprintf("could not decode contentChanges=%v", params.ContentChanges...))
	}

	myServer.glspServer.Log.Info(fmt.Sprintf("got Change: %+v", params.ContentChanges...))
	doc.ApplyContentChanges(changes.Text)
	myServer.refreshDiagnostics(doc, context.Notify, true)

	myServer.glspServer.Log.Info(fmt.Sprintf("new content=%s", doc.Content))
	myServer.glspServer.Log.Info(fmt.Sprintf("new content=%s", doc.Tree.RootNode()))

	return nil
}

func complete(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	requestPos := params.TextDocumentPositionParams.Position
	pos := document.DocumentPosition{Line: requestPos.Line, Char: requestPos.Character}
	methods, err := doc.GetMethodCompletions(pos)
	if err != nil {
		myServer.glspServer.Log.Error(fmt.Sprintf("completion error:=%s", err))
		return nil, err
	}

	completionList := make([]protocol.CompletionItem, len(methods))

	for i, c := range methods {
		completionList[i] = protocol.CompletionItem{Label: c.Name}
	}

	myServer.glspServer.Log.Infof(fmt.Sprintf("sending completion list=%v", completionList))

	return completionList, err
}

func TokenTypeToIndex(tokenType string) (int, error) {
	lookUp := make(map[string]int)

	lookUp["boolean"] = 15
	lookUp["let"] = 15
	lookUp["fn"] = 15
	lookUp["return"] = 15
	lookUp["if"] = 15
	lookUp["else"] = 15
	lookUp["string_literal"] = 18
	lookUp["identifier"] = 9
	lookUp["number"] = 19
	lookUp["parameter"] = 7

	index, ok := lookUp[tokenType]
	myServer.glspServer.Log.Info(fmt.Sprintf("sending lsp tokentype=%d for monkeytoken=%s", index, tokenType))
	if !ok {
		return -1, fmt.Errorf("could not find index for tokenType=%s", tokenType)
	}
	return index, nil
}

func TokenTypeToModifier(tokenType string) (int, bool) {
	lookUp := make(map[string]int)

	lookUp["function_name"] = 4
	lookUp["value_name"] = 2

	index, ok := lookUp[tokenType]
	myServer.glspServer.Log.Info(fmt.Sprintf("sending lsp tokenmodifier=%d for monkeytoken=%s", index, tokenType))
	if !ok {
		return -1, false
	}
	return index, true
}

func highlight(context *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	myServer.glspServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	hls, _ := doc.GetHighLights()

	data := []uint32{}
	for _, hl := range hls {
		data = append(data, uint32(hl.Line))
		data = append(data, uint32(hl.StartChar))
		data = append(data, uint32(hl.Length))

		tokenIndex, err := TokenTypeToIndex(hl.TokenType)
		if err != nil {
			return nil, fmt.Errorf("error in token lookup for tokenType=%s", hl.TokenType)
		}

		data = append(data, uint32(tokenIndex))
		tokenModifier, ok := TokenTypeToModifier(hl.TokenType)
		if !ok {
			tokenModifier = 0
		}
		data = append(data, uint32(tokenModifier))
	}

	return &protocol.SemanticTokens{
		Data: data,
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

	myServer.glspServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return nil, nil
}

func highLightRangeDelta(context *glsp.Context, params *protocol.SemanticTokensDeltaParams) (any, error) {
	myServer.glspServer.Log.Info(fmt.Sprintf("got token request for: %s", params.TextDocument.URI))
	return nil, nil
}

func AddTokenLegend(h *protocol.ServerCapabilities) {
	h.SemanticTokensProvider.(*protocol.SemanticTokensOptions).Legend = protocol.SemanticTokensLegend{
		TokenTypes: []string{
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

func (s *Server) refreshDiagnostics(doc *document.Document, notify glsp.NotifyFunc, delay bool) {
	if doc.NeedsReFreshDiagnostics {
		return
	}

	myServer.glspServer.Log.Info("calculation diagnostics")
	doc.NeedsReFreshDiagnostics = true

	go func() {

		if delay {
			time.Sleep(500 * time.Millisecond)
		}
		doc.NeedsReFreshDiagnostics = false

		diagnostics := []protocol.Diagnostic{}
		for _, d := range doc.GetDiagnostics() {

			var severity protocol.DiagnosticSeverity

			switch d.Severty {
			case document.ERROR:
				severity = protocol.DiagnosticSeverityError
			case document.HINT:
				severity = protocol.DiagnosticSeverityHint
			case document.WARNING:
				severity = protocol.DiagnosticSeverityWarning
			case document.INFORMATION:
				severity = protocol.DiagnosticSeverityInformation
			}

			pDiagnostic := protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: d.Start.Line, Character: d.Start.Char},
					End:   protocol.Position{Line: d.End.Line, Character: d.End.Char},
				},
				Severity: &severity,
				Message:  d.Message,
			}

			diagnostics = append(diagnostics, pDiagnostic)
		}

		myServer.glspServer.Log.Info(fmt.Sprintf("sending diagnostics for doc=%s", doc.Uri))

		go notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         doc.Uri,
			Diagnostics: diagnostics,
		})

	}()
}
