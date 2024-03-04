package document

type Position struct {
	Row    uint32
	Column uint32
}

type SemanticToken struct {
	Start Position
	End   Position
	Type  string
}
