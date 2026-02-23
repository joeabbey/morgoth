package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeabbey/morgoth/internal/lexer"
)

func parse(t *testing.T, input string) *Program {
	t.Helper()
	l := lexer.New(input)
	p := New(l)
	prog := p.Parse()
	for _, err := range p.Errors() {
		t.Errorf("parse error: %s", err)
	}
	return prog
}

func parseExpectErrors(input string) (*Program, []string) {
	l := lexer.New(input)
	p := New(l)
	prog := p.Parse()
	return prog, p.Errors()
}

// --- Simple statement tests ---

func TestLetStmt(t *testing.T) {
	prog := parse(t, `let x = 5;`)
	if len(prog.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(prog.Items))
	}
	stmt, ok := prog.Items[0].(*LetStmt)
	if !ok {
		t.Fatalf("expected *LetStmt, got %T", prog.Items[0])
	}
	if stmt.Name != "x" {
		t.Errorf("expected name x, got %s", stmt.Name)
	}
	lit, ok := stmt.Value.(*IntLitExpr)
	if !ok {
		t.Fatalf("expected *IntLitExpr, got %T", stmt.Value)
	}
	if lit.Value != 5 {
		t.Errorf("expected value 5, got %d", lit.Value)
	}
}

func TestLetStmtWithType(t *testing.T) {
	prog := parse(t, `let x: int = 42;`)
	stmt := prog.Items[0].(*LetStmt)
	if stmt.Name != "x" {
		t.Errorf("expected name x, got %s", stmt.Name)
	}
	if stmt.TypeAnnotation != "int" {
		t.Errorf("expected type int, got %s", stmt.TypeAnnotation)
	}
}

func TestConstStmt(t *testing.T) {
	prog := parse(t, `const y = 5;`)
	if len(prog.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(prog.Items))
	}
	stmt, ok := prog.Items[0].(*ConstStmt)
	if !ok {
		t.Fatalf("expected *ConstStmt, got %T", prog.Items[0])
	}
	if stmt.Name != "y" {
		t.Errorf("expected name y, got %s", stmt.Name)
	}
}

func TestReturnStmt(t *testing.T) {
	prog := parse(t, `return 42;`)
	stmt, ok := prog.Items[0].(*ReturnStmt)
	if !ok {
		t.Fatalf("expected *ReturnStmt, got %T", prog.Items[0])
	}
	lit, ok := stmt.Value.(*IntLitExpr)
	if !ok {
		t.Fatalf("expected *IntLitExpr, got %T", stmt.Value)
	}
	if lit.Value != 42 {
		t.Errorf("expected 42, got %d", lit.Value)
	}
}

func TestDecreeStmt(t *testing.T) {
	prog := parse(t, `decree "zero_indexed";`)
	stmt, ok := prog.Items[0].(*DecreeStmt)
	if !ok {
		t.Fatalf("expected *DecreeStmt, got %T", prog.Items[0])
	}
	if stmt.Value != "zero_indexed" {
		t.Errorf("expected zero_indexed, got %s", stmt.Value)
	}
}

// --- Expression precedence tests ---

func TestBinaryPrecedence(t *testing.T) {
	// 1 + 2 * 3 should parse as 1 + (2 * 3)
	prog := parse(t, `1 + 2 * 3;`)
	es := prog.Items[0].(*ExprStmt)
	bin, ok := es.Expression.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected *BinaryExpr, got %T", es.Expression)
	}
	if bin.Op != "+" {
		t.Errorf("expected +, got %s", bin.Op)
	}
	left, ok := bin.Left.(*IntLitExpr)
	if !ok {
		t.Fatalf("expected left *IntLitExpr, got %T", bin.Left)
	}
	if left.Value != 1 {
		t.Errorf("expected left=1, got %d", left.Value)
	}
	right, ok := bin.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected right *BinaryExpr, got %T", bin.Right)
	}
	if right.Op != "*" {
		t.Errorf("expected *, got %s", right.Op)
	}
}

func TestComparisonPrecedence(t *testing.T) {
	// a + b >= c should parse as (a + b) >= c
	prog := parse(t, `a + b >= c;`)
	es := prog.Items[0].(*ExprStmt)
	bin := es.Expression.(*BinaryExpr)
	if bin.Op != ">=" {
		t.Errorf("expected >=, got %s", bin.Op)
	}
	left := bin.Left.(*BinaryExpr)
	if left.Op != "+" {
		t.Errorf("expected left op +, got %s", left.Op)
	}
}

func TestLogicPrecedence(t *testing.T) {
	// a and b or c should parse as (a and b) or c
	prog := parse(t, `a and b or c;`)
	es := prog.Items[0].(*ExprStmt)
	bin := es.Expression.(*BinaryExpr)
	if bin.Op != "or" {
		t.Errorf("expected or at top level, got %s", bin.Op)
	}
	left := bin.Left.(*BinaryExpr)
	if left.Op != "and" {
		t.Errorf("expected and, got %s", left.Op)
	}
}

func TestUnaryExpr(t *testing.T) {
	prog := parse(t, `-5;`)
	es := prog.Items[0].(*ExprStmt)
	unary, ok := es.Expression.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected *UnaryExpr, got %T", es.Expression)
	}
	if unary.Op != "-" {
		t.Errorf("expected -, got %s", unary.Op)
	}
}

// --- Specific construct tests ---

func TestIfElseExpr(t *testing.T) {
	prog := parse(t, `if x { 1 } else { 2 };`)
	es := prog.Items[0].(*ExprStmt)
	ifExpr, ok := es.Expression.(*IfExpr)
	if !ok {
		t.Fatalf("expected *IfExpr, got %T", es.Expression)
	}
	if ifExpr.Then == nil {
		t.Fatal("expected Then block")
	}
	if ifExpr.Else == nil {
		t.Fatal("expected Else block")
	}
}

func TestIfElseIfExpr(t *testing.T) {
	prog := parse(t, `if a { 1 } else if b { 2 } else { 3 };`)
	es := prog.Items[0].(*ExprStmt)
	ifExpr := es.Expression.(*IfExpr)
	elseIf, ok := ifExpr.Else.(*IfExpr)
	if !ok {
		t.Fatalf("expected else-if chain, got %T", ifExpr.Else)
	}
	if elseIf.Else == nil {
		t.Fatal("expected final else block")
	}
}

func TestMatchExpr(t *testing.T) {
	input := `match x {
		1 => "one",
		2 => "two",
		_ => "other",
	};`
	prog := parse(t, input)
	es := prog.Items[0].(*ExprStmt)
	m, ok := es.Expression.(*MatchExpr)
	if !ok {
		t.Fatalf("expected *MatchExpr, got %T", es.Expression)
	}
	if len(m.Arms) != 3 {
		t.Fatalf("expected 3 arms, got %d", len(m.Arms))
	}
	// First arm: literal pattern
	_, ok = m.Arms[0].Pattern.(*LiteralPattern)
	if !ok {
		t.Errorf("expected LiteralPattern, got %T", m.Arms[0].Pattern)
	}
	// Last arm: wildcard
	_, ok = m.Arms[2].Pattern.(*WildcardPattern)
	if !ok {
		t.Errorf("expected WildcardPattern, got %T", m.Arms[2].Pattern)
	}
}

func TestMatchTypedPattern(t *testing.T) {
	input := `match x {
		n: int => n,
	};`
	prog := parse(t, input)
	es := prog.Items[0].(*ExprStmt)
	m := es.Expression.(*MatchExpr)
	tp, ok := m.Arms[0].Pattern.(*TypedPattern)
	if !ok {
		t.Fatalf("expected *TypedPattern, got %T", m.Arms[0].Pattern)
	}
	if tp.Name != "n" || tp.TypeName != "int" {
		t.Errorf("expected n: int, got %s: %s", tp.Name, tp.TypeName)
	}
}

func TestMatchGuardedPattern(t *testing.T) {
	input := `match x {
		n: int if n < 0 => doom("negative"),
	};`
	prog := parse(t, input)
	es := prog.Items[0].(*ExprStmt)
	m := es.Expression.(*MatchExpr)
	gp, ok := m.Arms[0].Pattern.(*GuardedPattern)
	if !ok {
		t.Fatalf("expected *GuardedPattern, got %T", m.Arms[0].Pattern)
	}
	inner, ok := gp.Inner.(*TypedPattern)
	if !ok {
		t.Fatalf("expected *TypedPattern inside guard, got %T", gp.Inner)
	}
	if inner.Name != "n" || inner.TypeName != "int" {
		t.Errorf("expected n: int, got %s: %s", inner.Name, inner.TypeName)
	}
}

func TestGuardExpr(t *testing.T) {
	input := `guard x >= 2 else doom("too small");`
	prog := parse(t, input)
	es := prog.Items[0].(*ExprStmt)
	g, ok := es.Expression.(*GuardExpr)
	if !ok {
		t.Fatalf("expected *GuardExpr, got %T", es.Expression)
	}
	if g.Condition == nil {
		t.Fatal("expected guard condition")
	}
	if g.ElseBody == nil {
		t.Fatal("expected guard else body")
	}
}

func TestOkErrExpr(t *testing.T) {
	prog := parse(t, `ok(42);`)
	es := prog.Items[0].(*ExprStmt)
	okExpr, ok := es.Expression.(*OkExpr)
	if !ok {
		t.Fatalf("expected *OkExpr, got %T", es.Expression)
	}
	lit := okExpr.Inner.(*IntLitExpr)
	if lit.Value != 42 {
		t.Errorf("expected 42, got %d", lit.Value)
	}

	prog2 := parse(t, `err("bad");`)
	es2 := prog2.Items[0].(*ExprStmt)
	errExpr, ok := es2.Expression.(*ErrExpr)
	if !ok {
		t.Fatalf("expected *ErrExpr, got %T", es2.Expression)
	}
	str := errExpr.Inner.(*StringLitExpr)
	if str.Value != "bad" {
		t.Errorf("expected bad, got %s", str.Value)
	}
}

func TestSpeakElseExpr(t *testing.T) {
	prog := parse(t, `speak "hello" else doom("fail");`)
	es := prog.Items[0].(*ExprStmt)
	sp, ok := es.Expression.(*SpeakExpr)
	if !ok {
		t.Fatalf("expected *SpeakExpr, got %T", es.Expression)
	}
	if sp.ElseBody == nil {
		t.Fatal("expected else body")
	}
}

func TestSpeakWithoutElse(t *testing.T) {
	prog := parse(t, `speak "hello";`)
	es := prog.Items[0].(*ExprStmt)
	sp, ok := es.Expression.(*SpeakExpr)
	if !ok {
		t.Fatalf("expected *SpeakExpr, got %T", es.Expression)
	}
	if sp.ElseBody != nil {
		t.Fatal("expected no else body")
	}
}

func TestSorryExpr(t *testing.T) {
	prog := parse(t, `sorry(y);`)
	es := prog.Items[0].(*ExprStmt)
	s, ok := es.Expression.(*SorryExpr)
	if !ok {
		t.Fatalf("expected *SorryExpr, got %T", es.Expression)
	}
	if s.Name != "y" {
		t.Errorf("expected y, got %s", s.Name)
	}
}

func TestDoomExpr(t *testing.T) {
	prog := parse(t, `doom("error");`)
	es := prog.Items[0].(*ExprStmt)
	d, ok := es.Expression.(*DoomExpr)
	if !ok {
		t.Fatalf("expected *DoomExpr, got %T", es.Expression)
	}
	str := d.Message.(*StringLitExpr)
	if str.Value != "error" {
		t.Errorf("expected error, got %s", str.Value)
	}
}

func TestChantExpr(t *testing.T) {
	prog := parse(t, `chant "stdio";`)
	es := prog.Items[0].(*ExprStmt)
	c, ok := es.Expression.(*ChantExpr)
	if !ok {
		t.Fatalf("expected *ChantExpr, got %T", es.Expression)
	}
	str := c.Name.(*StringLitExpr)
	if str.Value != "stdio" {
		t.Errorf("expected stdio, got %s", str.Value)
	}
}

func TestPropagateExpr(t *testing.T) {
	prog := parse(t, `x?;`)
	es := prog.Items[0].(*ExprStmt)
	prop, ok := es.Expression.(*PropagateExpr)
	if !ok {
		t.Fatalf("expected *PropagateExpr, got %T", es.Expression)
	}
	ident := prop.Inner.(*IdentExpr)
	if ident.Name != "x" {
		t.Errorf("expected x, got %s", ident.Name)
	}
}

func TestAsExpr(t *testing.T) {
	prog := parse(t, `s as int;`)
	es := prog.Items[0].(*ExprStmt)
	asExpr, ok := es.Expression.(*AsExpr)
	if !ok {
		t.Fatalf("expected *AsExpr, got %T", es.Expression)
	}
	if asExpr.TypeName != "int" {
		t.Errorf("expected int, got %s", asExpr.TypeName)
	}
}

func TestCallExpr(t *testing.T) {
	prog := parse(t, `foo(1, 2);`)
	es := prog.Items[0].(*ExprStmt)
	call, ok := es.Expression.(*CallExpr)
	if !ok {
		t.Fatalf("expected *CallExpr, got %T", es.Expression)
	}
	fn := call.Function.(*IdentExpr)
	if fn.Name != "foo" {
		t.Errorf("expected foo, got %s", fn.Name)
	}
	if len(call.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(call.Args))
	}
}

func TestIndexExpr(t *testing.T) {
	prog := parse(t, `xs[0];`)
	es := prog.Items[0].(*ExprStmt)
	idx, ok := es.Expression.(*IndexExpr)
	if !ok {
		t.Fatalf("expected *IndexExpr, got %T", es.Expression)
	}
	left := idx.Left.(*IdentExpr)
	if left.Name != "xs" {
		t.Errorf("expected xs, got %s", left.Name)
	}
}

func TestArrayLiteral(t *testing.T) {
	prog := parse(t, `[1, 2, 3];`)
	es := prog.Items[0].(*ExprStmt)
	arr, ok := es.Expression.(*ArrayLitExpr)
	if !ok {
		t.Fatalf("expected *ArrayLitExpr, got %T", es.Expression)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestMapLiteral(t *testing.T) {
	prog := parse(t, `{ "a": 1, "b": 2 };`)
	es := prog.Items[0].(*ExprStmt)
	m, ok := es.Expression.(*MapLitExpr)
	if !ok {
		t.Fatalf("expected *MapLitExpr, got %T", es.Expression)
	}
	if len(m.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(m.Pairs))
	}
}

func TestBlockExpr(t *testing.T) {
	// Block with statements and a final expression
	prog := parse(t, `{ let x = 1; x };`)
	es := prog.Items[0].(*ExprStmt)
	block, ok := es.Expression.(*BlockExpr)
	if !ok {
		t.Fatalf("expected *BlockExpr, got %T", es.Expression)
	}
	if len(block.Stmts) != 1 {
		t.Errorf("expected 1 stmt, got %d", len(block.Stmts))
	}
	if block.FinalExpr == nil {
		t.Fatal("expected final expr")
	}
}

func TestFnDecl(t *testing.T) {
	prog := parse(t, `fn add(a, b) { a + b }`)
	fn, ok := prog.Items[0].(*FnDecl)
	if !ok {
		t.Fatalf("expected *FnDecl, got %T", prog.Items[0])
	}
	if fn.Name != "add" {
		t.Errorf("expected add, got %s", fn.Name)
	}
	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}
	if fn.Body.FinalExpr == nil {
		t.Fatal("expected body final expr")
	}
}

func TestExternDecl(t *testing.T) {
	prog := parse(t, `extern fn malloc(n: int);`)
	ext, ok := prog.Items[0].(*ExternDecl)
	if !ok {
		t.Fatalf("expected *ExternDecl, got %T", prog.Items[0])
	}
	if ext.Name != "malloc" {
		t.Errorf("expected malloc, got %s", ext.Name)
	}
	if len(ext.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(ext.Params))
	}
	if ext.Params[0].Type != "int" {
		t.Errorf("expected param type int, got %s", ext.Params[0].Type)
	}
}

func TestAssignExpr(t *testing.T) {
	prog := parse(t, `x = 10;`)
	es := prog.Items[0].(*ExprStmt)
	assign, ok := es.Expression.(*AssignExpr)
	if !ok {
		t.Fatalf("expected *AssignExpr, got %T", es.Expression)
	}
	if assign.Name != "x" {
		t.Errorf("expected x, got %s", assign.Name)
	}
}

// --- Map vs Block disambiguation ---

func TestMapVsBlockDisambiguation(t *testing.T) {
	// Map: { "a": 1 }
	prog := parse(t, `{ "a": 1 };`)
	es := prog.Items[0].(*ExprStmt)
	_, isMap := es.Expression.(*MapLitExpr)
	if !isMap {
		t.Fatalf("expected map literal, got %T", es.Expression)
	}

	// Block: { 1 }
	prog2 := parse(t, `{ 1 };`)
	es2 := prog2.Items[0].(*ExprStmt)
	_, isBlock := es2.Expression.(*BlockExpr)
	if !isBlock {
		t.Fatalf("expected block, got %T", es2.Expression)
	}
}

// --- Example file tests ---

func TestExampleFiles(t *testing.T) {
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
			l := lexer.New(string(data))
			p := New(l)
			prog := p.Parse()
			if len(p.Errors()) > 0 {
				for _, e := range p.Errors() {
					t.Errorf("parse error: %s", e)
				}
			}
			if len(prog.Items) == 0 {
				t.Error("expected at least one item")
			}
		})
	}
}

// --- Specific example file structure tests ---

func TestHelloMorParsed(t *testing.T) {
	input := `let ok = chant "stdio";
speak "Hello, Morgoth!"
  else doom("stdout is cursed");`
	prog := parse(t, input)
	if len(prog.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(prog.Items))
	}
	// First: let ok = chant "stdio"
	letStmt, ok := prog.Items[0].(*LetStmt)
	if !ok {
		t.Fatalf("expected *LetStmt, got %T", prog.Items[0])
	}
	if letStmt.Name != "ok" {
		t.Errorf("expected name ok, got %s", letStmt.Name)
	}
	_, isChant := letStmt.Value.(*ChantExpr)
	if !isChant {
		t.Errorf("expected ChantExpr value, got %T", letStmt.Value)
	}

	// Second: speak ... else doom(...)
	exprStmt, ok := prog.Items[1].(*ExprStmt)
	if !ok {
		t.Fatalf("expected *ExprStmt, got %T", prog.Items[1])
	}
	speakExpr, ok := exprStmt.Expression.(*SpeakExpr)
	if !ok {
		t.Fatalf("expected *SpeakExpr, got %T", exprStmt.Expression)
	}
	if speakExpr.ElseBody == nil {
		t.Fatal("expected speak else body")
	}
}

func TestSpawnExpr(t *testing.T) {
	prog := parse(t, `spawn { speak "hi" };`)
	es := prog.Items[0].(*ExprStmt)
	sp, ok := es.Expression.(*SpawnExpr)
	if !ok {
		t.Fatalf("expected *SpawnExpr, got %T", es.Expression)
	}
	if sp.Body == nil {
		t.Fatal("expected spawn body")
	}
}

func TestAwaitAllExpr(t *testing.T) {
	prog := parse(t, `await_all();`)
	es := prog.Items[0].(*ExprStmt)
	_, ok := es.Expression.(*AwaitAllExpr)
	if !ok {
		t.Fatalf("expected *AwaitAllExpr, got %T", es.Expression)
	}
}

func TestResultMorParsed(t *testing.T) {
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
	prog := parse(t, input)
	if len(prog.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(prog.Items))
	}
	// fn parse_int
	fn1, ok := prog.Items[0].(*FnDecl)
	if !ok {
		t.Fatalf("expected *FnDecl, got %T", prog.Items[0])
	}
	if fn1.Name != "parse_int" {
		t.Errorf("expected parse_int, got %s", fn1.Name)
	}
	// fn read_number
	fn2, ok := prog.Items[1].(*FnDecl)
	if !ok {
		t.Fatalf("expected *FnDecl, got %T", prog.Items[1])
	}
	if fn2.Name != "read_number" {
		t.Errorf("expected read_number, got %s", fn2.Name)
	}
	// match expression
	exprStmt, ok := prog.Items[2].(*ExprStmt)
	if !ok {
		t.Fatalf("expected *ExprStmt, got %T", prog.Items[2])
	}
	_, isMatch := exprStmt.Expression.(*MatchExpr)
	if !isMatch {
		t.Fatalf("expected *MatchExpr, got %T", exprStmt.Expression)
	}
}
