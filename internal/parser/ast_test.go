package parser

import (
	"testing"

	"github.com/joeabbey/morgoth/internal/token"
)

func TestProgramTokenLiteral(t *testing.T) {
	prog := &Program{
		Items: []Item{
			&LetStmt{Token: token.Token{Type: token.LET, Literal: "let"}},
		},
	}
	if got := prog.TokenLiteral(); got != "let" {
		t.Errorf("Program.TokenLiteral() = %q, want %q", got, "let")
	}
}

func TestProgramTokenLiteralEmpty(t *testing.T) {
	prog := &Program{}
	if got := prog.TokenLiteral(); got != "" {
		t.Errorf("empty Program.TokenLiteral() = %q, want %q", got, "")
	}
}

func TestNodeTokenLiterals(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want string
	}{
		{"FnDecl", &FnDecl{Token: token.Token{Literal: "fn"}}, "fn"},
		{"ExternDecl", &ExternDecl{Token: token.Token{Literal: "extern"}}, "extern"},
		{"LetStmt", &LetStmt{Token: token.Token{Literal: "let"}}, "let"},
		{"ConstStmt", &ConstStmt{Token: token.Token{Literal: "const"}}, "const"},
		{"ReturnStmt", &ReturnStmt{Token: token.Token{Literal: "return"}}, "return"},
		{"DecreeStmt", &DecreeStmt{Token: token.Token{Literal: "decree"}}, "decree"},
		{"ExprStmt", &ExprStmt{Token: token.Token{Literal: "x"}}, "x"},
		{"IntLitExpr", &IntLitExpr{Token: token.Token{Literal: "42"}}, "42"},
		{"FloatLitExpr", &FloatLitExpr{Token: token.Token{Literal: "3.14"}}, "3.14"},
		{"StringLitExpr", &StringLitExpr{Token: token.Token{Literal: "hello"}}, "hello"},
		{"BoolLitExpr", &BoolLitExpr{Token: token.Token{Literal: "true"}}, "true"},
		{"NilLitExpr", &NilLitExpr{Token: token.Token{Literal: "nil"}}, "nil"},
		{"IdentExpr", &IdentExpr{Token: token.Token{Literal: "x"}}, "x"},
		{"ArrayLitExpr", &ArrayLitExpr{Token: token.Token{Literal: "["}}, "["},
		{"MapLitExpr", &MapLitExpr{Token: token.Token{Literal: "{"}}, "{"},
		{"BinaryExpr", &BinaryExpr{Token: token.Token{Literal: "+"}}, "+"},
		{"UnaryExpr", &UnaryExpr{Token: token.Token{Literal: "-"}}, "-"},
		{"AssignExpr", &AssignExpr{Token: token.Token{Literal: "="}}, "="},
		{"IndexAssignExpr", &IndexAssignExpr{Token: token.Token{Literal: "="}}, "="},
		{"DotAssignExpr", &DotAssignExpr{Token: token.Token{Literal: "="}}, "="},
		{"CallExpr", &CallExpr{Token: token.Token{Literal: "("}}, "("},
		{"IndexExpr", &IndexExpr{Token: token.Token{Literal: "["}}, "["},
		{"DotExpr", &DotExpr{Token: token.Token{Literal: "."}}, "."},
		{"PropagateExpr", &PropagateExpr{Token: token.Token{Literal: "?"}}, "?"},
		{"IfExpr", &IfExpr{Token: token.Token{Literal: "if"}}, "if"},
		{"MatchExpr", &MatchExpr{Token: token.Token{Literal: "match"}}, "match"},
		{"GuardExpr", &GuardExpr{Token: token.Token{Literal: "guard"}}, "guard"},
		{"BlockExpr", &BlockExpr{Token: token.Token{Literal: "{"}}, "{"},
		{"OkExpr", &OkExpr{Token: token.Token{Literal: "ok"}}, "ok"},
		{"ErrExpr", &ErrExpr{Token: token.Token{Literal: "err"}}, "err"},
		{"AsExpr", &AsExpr{Token: token.Token{Literal: "as"}}, "as"},
		{"SpeakExpr", &SpeakExpr{Token: token.Token{Literal: "speak"}}, "speak"},
		{"SorryExpr", &SorryExpr{Token: token.Token{Literal: "sorry"}}, "sorry"},
		{"DoomExpr", &DoomExpr{Token: token.Token{Literal: "doom"}}, "doom"},
		{"ChantExpr", &ChantExpr{Token: token.Token{Literal: "chant"}}, "chant"},
		{"FnLitExpr", &FnLitExpr{Token: token.Token{Literal: "fn"}}, "fn"},
		{"SpawnExpr", &SpawnExpr{Token: token.Token{Literal: "spawn"}}, "spawn"},
		{"AwaitAllExpr", &AwaitAllExpr{Token: token.Token{Literal: "await_all"}}, "await_all"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.TokenLiteral(); got != tt.want {
				t.Errorf("%s.TokenLiteral() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestPatternTokenLiterals(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		want    string
	}{
		{"WildcardPattern", &WildcardPattern{Token: token.Token{Literal: "_"}}, "_"},
		{"LiteralPattern", &LiteralPattern{Token: token.Token{Literal: "42"}}, "42"},
		{"IdentPattern", &IdentPattern{Token: token.Token{Literal: "x"}}, "x"},
		{"TypedPattern", &TypedPattern{Token: token.Token{Literal: "n"}}, "n"},
		{"GuardedPattern", &GuardedPattern{Token: token.Token{Literal: "if"}}, "if"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.TokenLiteral(); got != tt.want {
				t.Errorf("%s.TokenLiteral() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

// Verify interface compliance at compile time.
var (
	_ Item = (*FnDecl)(nil)
	_ Item = (*ExternDecl)(nil)
	_ Item = (*LetStmt)(nil)
	_ Item = (*ConstStmt)(nil)
	_ Item = (*ReturnStmt)(nil)
	_ Item = (*DecreeStmt)(nil)
	_ Item = (*ExprStmt)(nil)

	_ Stmt = (*LetStmt)(nil)
	_ Stmt = (*ConstStmt)(nil)
	_ Stmt = (*ReturnStmt)(nil)
	_ Stmt = (*DecreeStmt)(nil)
	_ Stmt = (*ExprStmt)(nil)

	_ Expr = (*IntLitExpr)(nil)
	_ Expr = (*FloatLitExpr)(nil)
	_ Expr = (*StringLitExpr)(nil)
	_ Expr = (*BoolLitExpr)(nil)
	_ Expr = (*NilLitExpr)(nil)
	_ Expr = (*IdentExpr)(nil)
	_ Expr = (*ArrayLitExpr)(nil)
	_ Expr = (*MapLitExpr)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*UnaryExpr)(nil)
	_ Expr = (*AssignExpr)(nil)
	_ Expr = (*IndexAssignExpr)(nil)
	_ Expr = (*DotAssignExpr)(nil)
	_ Expr = (*CallExpr)(nil)
	_ Expr = (*IndexExpr)(nil)
	_ Expr = (*DotExpr)(nil)
	_ Expr = (*PropagateExpr)(nil)
	_ Expr = (*IfExpr)(nil)
	_ Expr = (*MatchExpr)(nil)
	_ Expr = (*GuardExpr)(nil)
	_ Expr = (*BlockExpr)(nil)
	_ Expr = (*OkExpr)(nil)
	_ Expr = (*ErrExpr)(nil)
	_ Expr = (*AsExpr)(nil)
	_ Expr = (*SpeakExpr)(nil)
	_ Expr = (*SorryExpr)(nil)
	_ Expr = (*DoomExpr)(nil)
	_ Expr = (*ChantExpr)(nil)
	_ Expr = (*FnLitExpr)(nil)
	_ Expr = (*SpawnExpr)(nil)
	_ Expr = (*AwaitAllExpr)(nil)

	_ Pattern = (*WildcardPattern)(nil)
	_ Pattern = (*LiteralPattern)(nil)
	_ Pattern = (*IdentPattern)(nil)
	_ Pattern = (*TypedPattern)(nil)
	_ Pattern = (*GuardedPattern)(nil)
)
