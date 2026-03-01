package token

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tt   TokenType
		want string
	}{
		{INT, "INT"},
		{IDENT, "IDENT"},
		{FN, "FN"},
		{LPAREN, "LPAREN"},
		{EOF, "EOF"},
		{ILLEGAL, "ILLEGAL"},
		{SPEAK, "SPEAK"},
		{DECREE, "DECREE"},
		{SPAWN, "SPAWN"},
	}
	for _, tt := range tests {
		got := tt.tt.String()
		if got != tt.want {
			t.Errorf("TokenType(%d).String() = %q, want %q", int(tt.tt), got, tt.want)
		}
	}
}

func TestTokenTypeStringUnknown(t *testing.T) {
	unknown := TokenType(9999)
	got := unknown.String()
	if got == "" {
		t.Error("expected non-empty string for unknown TokenType")
	}
}

func TestLookupIdentKeywords(t *testing.T) {
	tests := []struct {
		ident string
		want  TokenType
	}{
		{"let", LET},
		{"const", CONST},
		{"fn", FN},
		{"return", RETURN},
		{"if", IF},
		{"else", ELSE},
		{"match", MATCH},
		{"guard", GUARD},
		{"doom", DOOM},
		{"ok", OK},
		{"err", ERR},
		{"nil", NIL},
		{"true", TRUE},
		{"false", FALSE},
		{"spawn", SPAWN},
		{"await_all", AWAIT_ALL},
		{"decree", DECREE},
		{"chant", CHANT},
		{"sorry", SORRY},
		{"speak", SPEAK},
		{"and", AND},
		{"or", OR},
		{"as", AS},
		{"ref", REF},
		{"extern", EXTERN},
	}
	for _, tt := range tests {
		got := LookupIdent(tt.ident)
		if got != tt.want {
			t.Errorf("LookupIdent(%q) = %v, want %v", tt.ident, got, tt.want)
		}
	}
}

func TestLookupIdentNonKeyword(t *testing.T) {
	tests := []string{"foo", "myVar", "x", "hello_world", "FN", "LET"}
	for _, ident := range tests {
		got := LookupIdent(ident)
		if got != IDENT {
			t.Errorf("LookupIdent(%q) = %v, want IDENT", ident, got)
		}
	}
}

func TestSemicolonTrigger(t *testing.T) {
	triggers := []TokenType{INT, FLOAT, STRING, IDENT, TRUE, FALSE, NIL, RPAREN, RBRACKET, RBRACE, QUESTION, OK, ERR}
	for _, tt := range triggers {
		if !SemicolonTrigger(tt) {
			t.Errorf("SemicolonTrigger(%v) = false, want true", tt)
		}
	}

	nonTriggers := []TokenType{PLUS, MINUS, STAR, LPAREN, LBRACKET, LBRACE, COMMA, COLON, FN, LET, IF, SPEAK, EOF}
	for _, tt := range nonTriggers {
		if SemicolonTrigger(tt) {
			t.Errorf("SemicolonTrigger(%v) = true, want false", tt)
		}
	}
}

func TestStartsStatement(t *testing.T) {
	starters := []TokenType{LET, CONST, FN, MATCH, IF, GUARD, RETURN, DECREE, SPAWN, SPEAK, DOOM, SORRY, CHANT}
	for _, tt := range starters {
		if !StartsStatement(tt) {
			t.Errorf("StartsStatement(%v) = false, want true", tt)
		}
	}

	nonStarters := []TokenType{INT, IDENT, PLUS, LPAREN, ELSE, TRUE, FALSE, NIL, EOF, OK, ERR}
	for _, tt := range nonStarters {
		if StartsStatement(tt) {
			t.Errorf("StartsStatement(%v) = true, want false", tt)
		}
	}
}
