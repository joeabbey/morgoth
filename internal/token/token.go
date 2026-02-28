package token

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Literals
	INT TokenType = iota
	FLOAT
	STRING

	// Identifiers
	IDENT

	// Keywords
	LET
	CONST
	FN
	RETURN
	IF
	ELSE
	MATCH
	GUARD
	DOOM
	OK
	ERR
	NIL
	TRUE
	FALSE
	REF
	EXTERN
	SPAWN
	AWAIT_ALL
	DECREE
	CHANT
	SORRY
	SPEAK
	AND
	OR
	AS

	// Operators
	PLUS      // +
	MINUS     // -
	STAR      // *
	SLASH     // /
	PERCENT   // %
	ASSIGN    // =
	EQ        // ==
	STRICT_EQ // ===
	NEQ       // !=
	LT        // <
	GT        // >
	LTE       // <=
	GTE       // >=
	BANG      // !
	AMP       // &

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACKET  // [
	RBRACKET  // ]
	LBRACE    // {
	RBRACE    // }
	COMMA     // ,
	SEMICOLON // ;
	COLON     // :
	ARROW     // =>
	DOT       // .
	QUESTION  // ?

	// Special
	EOF
	ILLEGAL
)

var tokenNames = map[TokenType]string{
	INT:       "INT",
	FLOAT:     "FLOAT",
	STRING:    "STRING",
	IDENT:     "IDENT",
	LET:       "LET",
	CONST:     "CONST",
	FN:        "FN",
	RETURN:    "RETURN",
	IF:        "IF",
	ELSE:      "ELSE",
	MATCH:     "MATCH",
	GUARD:     "GUARD",
	DOOM:      "DOOM",
	OK:        "OK",
	ERR:       "ERR",
	NIL:       "NIL",
	TRUE:      "TRUE",
	FALSE:     "FALSE",
	REF:       "REF",
	EXTERN:    "EXTERN",
	SPAWN:     "SPAWN",
	AWAIT_ALL: "AWAIT_ALL",
	DECREE:    "DECREE",
	CHANT:     "CHANT",
	SORRY:     "SORRY",
	SPEAK:     "SPEAK",
	AND:       "AND",
	OR:        "OR",
	AS:        "AS",
	PLUS:      "PLUS",
	MINUS:     "MINUS",
	STAR:      "STAR",
	SLASH:     "SLASH",
	PERCENT:   "PERCENT",
	ASSIGN:    "ASSIGN",
	EQ:        "EQ",
	STRICT_EQ: "STRICT_EQ",
	NEQ:       "NEQ",
	LT:        "LT",
	GT:        "GT",
	LTE:       "LTE",
	GTE:       "GTE",
	BANG:      "BANG",
	AMP:       "AMP",
	LPAREN:    "LPAREN",
	RPAREN:    "RPAREN",
	LBRACKET:  "LBRACKET",
	RBRACKET:  "RBRACKET",
	LBRACE:    "LBRACE",
	RBRACE:    "RBRACE",
	COMMA:     "COMMA",
	SEMICOLON: "SEMICOLON",
	COLON:     "COLON",
	ARROW:     "ARROW",
	DOT:       "DOT",
	QUESTION:  "QUESTION",
	EOF:       "EOF",
	ILLEGAL:   "ILLEGAL",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", int(t))
}

// Token represents a single lexical token with position information.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

var keywords = map[string]TokenType{
	"let":       LET,
	"const":     CONST,
	"fn":        FN,
	"return":    RETURN,
	"if":        IF,
	"else":      ELSE,
	"match":     MATCH,
	"guard":     GUARD,
	"doom":      DOOM,
	"ok":        OK,
	"err":       ERR,
	"nil":       NIL,
	"true":      TRUE,
	"false":     FALSE,
	"ref":       REF,
	"extern":    EXTERN,
	"spawn":     SPAWN,
	"await_all": AWAIT_ALL,
	"decree":    DECREE,
	"chant":     CHANT,
	"sorry":     SORRY,
	"speak":     SPEAK,
	"and":       AND,
	"or":        OR,
	"as":        AS,
}

// LookupIdent returns the TokenType for a given identifier string.
// If the identifier is a keyword, the corresponding keyword token type is returned.
// Otherwise, IDENT is returned.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// SemicolonTrigger returns true if a token of this type at the end of a line
// can trigger automatic semicolon insertion.
func SemicolonTrigger(t TokenType) bool {
	switch t {
	case INT, FLOAT, STRING, IDENT, TRUE, FALSE, NIL, RPAREN, RBRACKET, RBRACE, QUESTION, OK, ERR:
		return true
	}
	return false
}

// StartsStatement returns true if this token type is one of the keywords
// that can begin a new statement (used for semicolon insertion).
var statementStarters = map[TokenType]bool{
	LET:    true,
	CONST:  true,
	FN:     true,
	MATCH:  true,
	IF:     true,
	GUARD:  true,
	RETURN: true,
	DECREE: true,
	SPAWN:  true,
}

func StartsStatement(t TokenType) bool {
	return statementStarters[t]
}
