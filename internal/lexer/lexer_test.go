package lexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeabbey/morgoth/internal/token"
)

func TestSimpleTokens(t *testing.T) {
	input := `+ - * / % = == === != < > <= >= ! & ( ) [ ] { } , ; : => . ?`
	expected := []token.TokenType{
		token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
		token.ASSIGN, token.EQ, token.STRICT_EQ, token.NEQ,
		token.LT, token.GT, token.LTE, token.GTE,
		token.BANG, token.AMP,
		token.LPAREN, token.RPAREN, token.LBRACKET, token.RBRACKET,
		token.LBRACE, token.RBRACE,
		token.COMMA, token.SEMICOLON, token.COLON, token.ARROW, token.DOT, token.QUESTION,
		token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("token[%d]: expected %s, got %s (literal=%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `let const fn return if else match guard doom ok err nil true false ref extern spawn await_all decree chant sorry speak and or as`
	expected := []token.TokenType{
		token.LET, token.CONST, token.FN, token.RETURN, token.IF, token.ELSE,
		token.MATCH, token.GUARD, token.DOOM, token.OK, token.ERR, token.NIL,
		token.TRUE, token.FALSE, token.REF, token.EXTERN, token.SPAWN,
		token.AWAIT_ALL, token.DECREE, token.CHANT, token.SORRY, token.SPEAK,
		token.AND, token.OR, token.AS,
		token.EOF,
	}
	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("keyword[%d]: expected %s, got %s (literal=%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"42", "42"},
		{"1_000", "1_000"},
		{"0xDEAD_BEEF", "0xDEAD_BEEF"},
		{"0xFF", "0xFF"},
		{"0x0", "0x0"},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.INT {
			t.Errorf("input %q: expected INT, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.literal {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.literal, tok.Literal)
		}
	}
}

func TestFloatLiterals(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"1_000.5", "1_000.5"},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.FLOAT {
			t.Errorf("input %q: expected FLOAT, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.literal {
			t.Errorf("input %q: expected literal %q, got %q", tt.input, tt.literal, tok.Literal)
		}
	}
}

func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`""`, ""},
		{`"line\nbreak"`, "line\nbreak"},
		{`"tab\there"`, "tab\there"},
		{`"null\0byte"`, "null\x00byte"},
		{`"escaped\"quote"`, `escaped"quote`},
		{`"back\\slash"`, `back\slash`},
	}
	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.STRING {
			t.Errorf("input %s: expected STRING, got %s", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input %s: expected literal %q, got %q", tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestLineComments(t *testing.T) {
	input := `let x = 5 # this is a comment
let y = 10`
	l := New(input)
	tokens := l.Tokenize()

	// Should see: LET IDENT ASSIGN INT SEMICOLON LET IDENT ASSIGN INT EOF
	// (semicolon inserted because INT before newline, LET starts next line)
	expectedTypes := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		token.SEMICOLON,
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		token.SEMICOLON, // trailing semicolon at EOF since last token is INT
		token.EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokenTypes(tokens))
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}

func TestBlockComments(t *testing.T) {
	input := `let x = #{ block comment }# 42`
	l := New(input)
	tokens := l.Tokenize()

	expectedTypes := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		token.SEMICOLON, // trailing EOF semicolon
		token.EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokenTypes(tokens))
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}

func TestNestedBlockComments(t *testing.T) {
	input := `let x = #{ outer #{ inner }# still comment }# 42`
	l := New(input)
	tokens := l.Tokenize()

	expectedTypes := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		token.SEMICOLON,
		token.EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokenTypes(tokens))
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}

func TestSemicolonInsertion(t *testing.T) {
	// Semicolons should be inserted at newlines when:
	// - Line ends with literal/ident/)/]/}
	// - Next line starts with let/const/fn/match/if/guard/return or EOF
	input := `let x = 5
let y = 10
fn foo() {
  return x
}`
	l := New(input)
	tokens := l.Tokenize()

	expectedTypes := []token.TokenType{
		// let x = 5
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		// ; (semicolon inserted: INT before newline, LET on next line)
		token.SEMICOLON,
		// let y = 10
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		// ; (semicolon inserted: INT before newline, FN on next line)
		token.SEMICOLON,
		// fn foo() {
		token.FN, token.IDENT, token.LPAREN, token.RPAREN, token.LBRACE,
		// return x
		token.RETURN, token.IDENT,
		// No semicolon here: IDENT before newline, but } is not a statement starter
		// }
		token.RBRACE,
		// ; (semicolon inserted: } before EOF)
		token.SEMICOLON,
		token.EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokenTypes(tokens))
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s (literal=%q)", i, exp, tokens[i].Type, tokens[i].Literal)
		}
	}
}

func TestSemicolonInsertionAtEOF(t *testing.T) {
	input := `let x = 5`
	l := New(input)
	tokens := l.Tokenize()

	// INT at end triggers semicolon before EOF
	expectedTypes := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT,
		token.SEMICOLON,
		token.EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokenTypes(tokens))
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}

func TestNoSemicolonInsertionMidExpression(t *testing.T) {
	// No semicolon inserted when the next line doesn't start with a statement keyword.
	input := `let x = 5 +
10`
	l := New(input)
	tokens := l.Tokenize()

	expectedTypes := []token.TokenType{
		token.LET, token.IDENT, token.ASSIGN, token.INT, token.PLUS, token.INT,
		token.SEMICOLON, // trailing at EOF
		token.EOF,
	}
	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expectedTypes), len(tokens), tokenTypes(tokens))
	}
	for i, exp := range expectedTypes {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}

func TestOperatorLongestMatch(t *testing.T) {
	// === should be parsed as one token, not = + ==
	input := `a === b`
	l := New(input)
	tokens := l.Tokenize()
	// IDENT STRICT_EQ IDENT SEMICOLON EOF
	if tokens[1].Type != token.STRICT_EQ {
		t.Errorf("expected STRICT_EQ, got %s (literal=%q)", tokens[1].Type, tokens[1].Literal)
	}

	// == should not consume a third =
	input2 := `a == b`
	l2 := New(input2)
	tokens2 := l2.Tokenize()
	if tokens2[1].Type != token.EQ {
		t.Errorf("expected EQ, got %s (literal=%q)", tokens2[1].Type, tokens2[1].Literal)
	}

	// => should be ARROW
	input3 := `a => b`
	l3 := New(input3)
	tokens3 := l3.Tokenize()
	if tokens3[1].Type != token.ARROW {
		t.Errorf("expected ARROW, got %s", tokens3[1].Type)
	}
}

func TestPositionTracking(t *testing.T) {
	input := "let x = 5"
	l := New(input)
	tok := l.NextToken() // let
	if tok.Line != 1 || tok.Col != 1 {
		t.Errorf("let: expected line=1 col=1, got line=%d col=%d", tok.Line, tok.Col)
	}
}

func TestExampleFilesNoIllegal(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("could not read examples dir: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".mor" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(examplesDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("could not read %s: %v", path, err)
			}
			l := New(string(data))
			tokens := l.Tokenize()
			for i, tok := range tokens {
				if tok.Type == token.ILLEGAL {
					t.Errorf("ILLEGAL token at index %d, line %d col %d: %q",
						i, tok.Line, tok.Col, tok.Literal)
				}
			}
		})
	}
}

func TestHelloMor(t *testing.T) {
	input := `# examples/hello.mor
let ok = chant "stdio";
speak "Hello, Morgoth!"
  else doom("stdout is cursed");`
	l := New(input)
	tokens := l.Tokenize()

	// Verify no ILLEGAL tokens and some key tokens present
	for i, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			t.Errorf("ILLEGAL at index %d: %q", i, tok.Literal)
		}
	}

	// Check key token sequence at start
	expected := []token.TokenType{
		token.LET, token.OK, token.ASSIGN, token.CHANT, token.STRING, token.SEMICOLON,
	}
	for i, exp := range expected {
		if i >= len(tokens) {
			t.Fatalf("ran out of tokens at index %d", i)
		}
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s (literal=%q)", i, exp, tokens[i].Type, tokens[i].Literal)
		}
	}
}

func TestMatchMor(t *testing.T) {
	input := `match token {
  "{" => speak "open",
  "}" => speak "close",
  _ => speak "unknown",
}`
	l := New(input)
	tokens := l.Tokenize()
	for i, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			t.Errorf("ILLEGAL at index %d, line %d: %q", i, tok.Line, tok.Literal)
		}
	}
}

func TestResultMor(t *testing.T) {
	input := `fn parse_int(s) {
  if s == "" { err("empty") }
  else ok(s as int)
}

fn read_number() {
  let n = parse_int("42")?;
  ok(n + 1)
}

match read_number() {
  ok(v) => speak v,
  err(e) => speak "error: " + e,
}`
	l := New(input)
	tokens := l.Tokenize()
	for i, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			t.Errorf("ILLEGAL at index %d, line %d col %d: %q", i, tok.Line, tok.Col, tok.Literal)
		}
	}

	// Check that ? is properly lexed
	found := false
	for _, tok := range tokens {
		if tok.Type == token.QUESTION {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected QUESTION token for ? operator")
	}

	// Check that 'as' is properly lexed as keyword
	foundAs := false
	for _, tok := range tokens {
		if tok.Type == token.AS {
			foundAs = true
			break
		}
	}
	if !foundAs {
		t.Error("expected AS token for 'as' keyword")
	}
}

func TestGuardMor(t *testing.T) {
	input := `fn main(args) {
  guard len(args) >= 2 else doom("usage: app <name>");
  let name = args[1];
  speak "Hello, " + name else doom("failed");
}

main(["app", "Sam"]);`
	l := New(input)
	tokens := l.Tokenize()
	for i, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			t.Errorf("ILLEGAL at index %d, line %d col %d: %q", i, tok.Line, tok.Col, tok.Literal)
		}
	}
}

func TestBlockCommentNestingCap(t *testing.T) {
	// Depth 3 nesting: the innermost #{ should NOT increase depth beyond 2,
	// so the first }# closes depth 2->1, second }# closes depth 1->0.
	// After the comment closes, "done" becomes visible as an IDENT token,
	// proving the cap worked (without cap, "done" would be inside the comment).
	input := `let x = #{ outer #{ inner #{ too deep }# still inner }# done }# 42`
	l := New(input)
	tokens := l.Tokenize()

	// With cap at depth 2, the comment closes early and "done" is visible.
	// (The trailing }# becomes } then # which starts a line comment eating 42.)
	foundDone := false
	for _, tok := range tokens {
		if tok.Type == token.IDENT && tok.Literal == "done" {
			foundDone = true
		}
	}
	if !foundDone {
		types := tokenTypes(tokens)
		t.Errorf("expected to find IDENT 'done' after capped block comment, got tokens: %v", types)
	}
}

func tokenTypes(tokens []token.Token) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t.Type.String()
	}
	return out
}
