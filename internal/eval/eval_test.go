package eval

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joeabbey/morgoth/internal/lexer"
	"github.com/joeabbey/morgoth/internal/parser"
)

// helper runs source through lex->parse->eval and returns the captured output and eval result.
func evalSource(t *testing.T, source string) (stdout string, result *Value, err error) {
	t.Helper()
	l := lexer.New(source)
	p := parser.New(l)
	prog := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	var buf bytes.Buffer
	ev := New()
	ev.SetOutput(&buf)
	result, err = ev.Eval(prog)
	return buf.String(), result, err
}

// --- Arithmetic ---

func TestArithmetic(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{`speak 1 + 2;`, "3\n"},
		{`speak 10 - 3;`, "7\n"},
		{`speak 4 * 5;`, "20\n"},
		{`speak 10 / 3;`, "3\n"},
		{`speak 10 % 3;`, "1\n"},
		{`speak -5;`, "-5\n"},
	}
	for _, tt := range tests {
		out, _, err := evalSource(t, tt.source)
		if err != nil {
			t.Errorf("source %q: unexpected error: %v", tt.source, err)
			continue
		}
		if out != tt.want {
			t.Errorf("source %q: got %q, want %q", tt.source, out, tt.want)
		}
	}
}

// --- String concatenation ---

func TestStringConcat(t *testing.T) {
	out, _, err := evalSource(t, `speak "hello" + " " + "world";`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello world\n" {
		t.Errorf("got %q, want %q", out, "hello world\n")
	}
}

// --- Comparisons ---

func TestComparisons(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{`speak 1 == 1;`, "true\n"},
		{`speak 1 == 2;`, "false\n"},
		{`speak 1 != 2;`, "true\n"},
		{`speak 3 < 5;`, "true\n"},
		{`speak 3 > 5;`, "false\n"},
		{`speak 3 >= 3;`, "true\n"},
		{`speak 3 <= 2;`, "false\n"},
	}
	for _, tt := range tests {
		out, _, err := evalSource(t, tt.source)
		if err != nil {
			t.Errorf("source %q: unexpected error: %v", tt.source, err)
			continue
		}
		if out != tt.want {
			t.Errorf("source %q: got %q, want %q", tt.source, out, tt.want)
		}
	}
}

// --- Let / Const / Sorry ---

func TestLetConst(t *testing.T) {
	out, _, err := evalSource(t, `
let x = 10;
x = 20;
speak x;
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "20\n" {
		t.Errorf("got %q, want %q", out, "20\n")
	}
}

func TestConstImmutable(t *testing.T) {
	_, _, err := evalSource(t, `
const x = 10;
x = 20;
`)
	if err == nil {
		t.Fatal("expected error when reassigning const")
	}
}

func TestSorryForgives(t *testing.T) {
	out, _, err := evalSource(t, `
const y = 5;
sorry(y);
y = 6;
speak y;
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "6\n" {
		t.Errorf("got %q, want %q", out, "6\n")
	}
}

// --- If/Else ---

func TestIfElse(t *testing.T) {
	out, _, err := evalSource(t, `
let x = if true { 1 } else { 2 };
speak x;
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "1\n" {
		t.Errorf("got %q, want %q", out, "1\n")
	}
}

func TestIfElseFalsy(t *testing.T) {
	out, _, err := evalSource(t, `
let x = if false { 1 } else { 2 };
speak x;
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "2\n" {
		t.Errorf("got %q, want %q", out, "2\n")
	}
}

// --- Functions ---

func TestFunctionCall(t *testing.T) {
	out, _, err := evalSource(t, `
fn add(a, b) { a + b }
speak add(3, 4);
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "7\n" {
		t.Errorf("got %q, want %q", out, "7\n")
	}
}

func TestFunctionReturn(t *testing.T) {
	out, _, err := evalSource(t, `
fn early(x) {
  if x > 0 { return x; }
  0
}
speak early(5);
speak early(-1);
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "5\n0\n" {
		t.Errorf("got %q, want %q", out, "5\n0\n")
	}
}

// --- Guard ---

func TestGuardPass(t *testing.T) {
	out, _, err := evalSource(t, `
fn greet(name) {
  guard name != "" else doom("empty name");
  speak "hi " + name;
}
greet("Sam");
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "hi Sam\n" {
		t.Errorf("got %q, want %q", out, "hi Sam\n")
	}
}

func TestGuardFail(t *testing.T) {
	_, _, err := evalSource(t, `
fn greet(name) {
  guard name != "" else doom("empty name");
  speak "hi " + name;
}
greet("");
`)
	if err == nil {
		t.Fatal("expected doom error from failed guard")
	}
	if doomErr, ok := err.(*DoomError); ok {
		if doomErr.Message != "empty name" {
			t.Errorf("got doom message %q, want %q", doomErr.Message, "empty name")
		}
	}
}

// --- Ok / Err / ? propagation ---

func TestOkErr(t *testing.T) {
	out, _, err := evalSource(t, `
let x = ok(42);
let y = err("bad");
speak x;
speak y;
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok(42)\nerr(bad)\n" {
		t.Errorf("got %q, want %q", out, "ok(42)\nerr(bad)\n")
	}
}

func TestPropagate(t *testing.T) {
	out, _, err := evalSource(t, `
fn get_val() { ok(10) }
fn use_val() {
  let v = get_val()?;
  ok(v + 1)
}
match use_val() {
  ok(v) => speak v,
  err(e) => speak e,
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "11\n" {
		t.Errorf("got %q, want %q", out, "11\n")
	}
}

func TestPropagateErr(t *testing.T) {
	out, _, err := evalSource(t, `
fn fail() { err("oops") }
fn use() {
  let v = fail()?;
  ok(v)
}
match use() {
  ok(v) => speak v,
  err(e) => speak "caught: " + e,
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "caught: oops\n" {
		t.Errorf("got %q, want %q", out, "caught: oops\n")
	}
}

// --- Match ---

func TestMatchLiteral(t *testing.T) {
	out, _, err := evalSource(t, `
let x = 2;
match x {
  1 => speak "one",
  2 => speak "two",
  _ => speak "other",
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "two\n" {
		t.Errorf("got %q, want %q", out, "two\n")
	}
}

func TestMatchWildcard(t *testing.T) {
	out, _, err := evalSource(t, `
match 99 {
  _ => speak "anything",
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "anything\n" {
		t.Errorf("got %q, want %q", out, "anything\n")
	}
}

func TestMatchIdent(t *testing.T) {
	out, _, err := evalSource(t, `
match 42 {
  n => speak n,
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "42\n" {
		t.Errorf("got %q, want %q", out, "42\n")
	}
}

// --- Arrays with decree ---

func TestArrayZeroIndexed(t *testing.T) {
	out, _, err := evalSource(t, `
let xs = [10, 20, 30];
decree "zero_indexed";
speak xs[0];
speak xs[2];
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "10\n30\n" {
		t.Errorf("got %q, want %q", out, "10\n30\n")
	}
}

func TestArrayOneIndexed(t *testing.T) {
	out, _, err := evalSource(t, `
let xs = [10, 20, 30];
decree "one_indexed";
speak xs[1];
speak xs[3];
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "10\n30\n" {
		t.Errorf("got %q, want %q", out, "10\n30\n")
	}
}

// --- Maps ---

func TestMapAccess(t *testing.T) {
	out, _, err := evalSource(t, `
decree "deterministic_hashing";
let m = { "a": 1, "b": 2 };
speak m["a"];
speak m["b"];
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "1\n2\n" {
		t.Errorf("got %q, want %q", out, "1\n2\n")
	}
}

// --- Example file golden tests ---

func TestExampleHello(t *testing.T)   { testExampleFile(t, "hello.mor") }
func TestExampleArrays(t *testing.T)  { testExampleFile(t, "arrays.mor") }
func TestExampleDecrees(t *testing.T) { testExampleFile(t, "decrees.mor") }
func TestExampleGuard(t *testing.T)   { testExampleFile(t, "guard.mor") }
func TestExampleMatch(t *testing.T)   { testExampleFile(t, "match.mor") }
func TestExampleResult(t *testing.T)  { testExampleFile(t, "result.mor") }

func TestGuardPropagatesElseValue(t *testing.T) {
	// Guard should cause non-local return from enclosing function with else value.
	out, _, err := evalSource(t, `
fn check(x) {
  guard x > 0 else "must be positive"
  ok(x)
}
speak check(-1);
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "must be positive\n" {
		t.Errorf("got %q, want %q", out, "must be positive\n")
	}
}

func TestGuardReturnFromFunction(t *testing.T) {
	out, _, err := evalSource(t, `
fn check(x) {
  guard x > 0 else err("must be positive")
  ok(x)
}
match check(-1) {
  ok(v) => speak v,
  err(e) => speak "error: " + e,
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "error: must be positive\n" {
		t.Errorf("got %q, want %q", out, "error: must be positive\n")
	}
}

func TestGuardWithDoomStillDooms(t *testing.T) {
	_, _, err := evalSource(t, `
fn check(x) {
  guard x > 0 else doom("negative")
  ok(x)
}
check(-1);
`)
	if err == nil {
		t.Fatal("expected doom error")
	}
	doomErr, ok := err.(*DoomError)
	if !ok {
		t.Fatalf("expected *DoomError, got %T: %v", err, err)
	}
	if doomErr.Message != "negative" {
		t.Errorf("got %q, want %q", doomErr.Message, "negative")
	}
}

func TestPropagateAtTopLevelDooms(t *testing.T) {
	_, _, err := evalSource(t, `err("oops")?;`)
	if err == nil {
		t.Fatal("expected error from ? at top level")
	}
	doomErr, ok := err.(*DoomError)
	if !ok {
		t.Fatalf("expected *DoomError, got %T: %v", err, err)
	}
	if !strings.Contains(doomErr.Message, "oops") {
		t.Errorf("expected doom message to contain 'oops', got %q", doomErr.Message)
	}
}

func TestSorrySameScope(t *testing.T) {
	out, _, err := evalSource(t, `
const x = 5;
sorry(x);
x = 6;
speak x;
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "6\n" {
		t.Errorf("got %q, want %q", out, "6\n")
	}
}

func TestSorryCrossScope(t *testing.T) {
	out, _, err := evalSource(t, `
const x = 5;
fn forgive_x() {
  sorry(x)
}
let result = forgive_x();
speak result;
`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "err(") {
		t.Errorf("expected err result from sorry in wrong scope, got %q", out)
	}
}

func TestChantEvaluatesArgument(t *testing.T) {
	_, _, err := evalSource(t, `chant doom("should doom");`)
	if err == nil {
		t.Fatal("expected doom error from chant argument evaluation")
	}
}

// --- Index assignment ---

func TestArrayIndexAssignment(t *testing.T) {
	out, _, err := evalSource(t, `
decree "zero_indexed"
let xs = [10, 20, 30]
xs[1] = 99
speak xs[1]
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "99\n" {
		t.Errorf("got %q, want %q", out, "99\n")
	}
}

func TestMapIndexAssignment(t *testing.T) {
	out, _, err := evalSource(t, `
decree "deterministic_hashing"
let m = { "a": 1, "b": 2 }
m["a"] = 42
speak m["a"]
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "42\n" {
		t.Errorf("got %q, want %q", out, "42\n")
	}
}

func TestMapDotAssignment(t *testing.T) {
	out, _, err := evalSource(t, `
decree "deterministic_hashing"
let m = { "x": 1 }
m.x = 100
speak m.x
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "100\n" {
		t.Errorf("got %q, want %q", out, "100\n")
	}
}

func TestMapAddNewKey(t *testing.T) {
	out, _, err := evalSource(t, `
decree "deterministic_hashing"
let m = { "a": 1 }
m["b"] = 2
speak m["b"]
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "2\n" {
		t.Errorf("got %q, want %q", out, "2\n")
	}
}

func TestArrayIndexAssignOutOfBounds(t *testing.T) {
	_, _, err := evalSource(t, `
decree "zero_indexed"
let xs = [1, 2, 3]
xs[5] = 99
`)
	if err == nil {
		t.Fatal("expected out of bounds error")
	}
}

func TestStrictEqualDifferentTypes(t *testing.T) {
	// === should be false for different types even if values look similar
	out, _, err := evalSource(t, `speak 1 === 1;`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "true\n" {
		t.Errorf("got %q, want %q", out, "true\n")
	}

	// ok(1) === ok(1) should be true with strict equal
	out2, _, err := evalSource(t, `speak ok(1) === ok(1);`)
	if err != nil {
		t.Fatal(err)
	}
	if out2 != "true\n" {
		t.Errorf("got %q, want %q", out2, "true\n")
	}

	// ok(1) === err(1) should be false
	out3, _, err := evalSource(t, `speak ok(1) === err(1);`)
	if err != nil {
		t.Fatal(err)
	}
	if out3 != "false\n" {
		t.Errorf("got %q, want %q", out3, "false\n")
	}
}

func TestAmbitiousMode(t *testing.T) {
	out, _, err := evalSource(t, `
decree "ambitious_mode"
let x = 5
x == 10
speak x
`)
	if err != nil {
		t.Fatal(err)
	}
	// In ambitious_mode, x == 10 should assign 10 to x (since 10 is truthy)
	if out != "10\n" {
		t.Errorf("got %q, want %q", out, "10\n")
	}
}

func TestAmbitiousModeNoAssignFalsy(t *testing.T) {
	out, _, err := evalSource(t, `
decree "ambitious_mode"
let x = 5
x == 0
speak x
`)
	if err != nil {
		t.Fatal(err)
	}
	// 0 is falsy, so == should compare normally (returns false), x stays 5
	if out != "5\n" {
		t.Errorf("got %q, want %q", out, "5\n")
	}
}

func TestDefaultIndexingIsWeekday(t *testing.T) {
	// Without any decree, indexing should be "weekday" mode
	// We can't test the exact behavior deterministically (depends on day),
	// but we can verify that after decree "zero_indexed", index 0 works
	out, _, err := evalSource(t, `
decree "zero_indexed"
let xs = [10, 20, 30]
speak xs[0]
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "10\n" {
		t.Errorf("got %q, want %q", out, "10\n")
	}
}

func TestReturnAtTopLevel(t *testing.T) {
	_, _, err := evalSource(t, `return 42;`)
	if err == nil {
		t.Fatal("expected error from return at top level")
	}
	doomErr, ok := err.(*DoomError)
	if !ok {
		t.Fatalf("expected *DoomError, got %T: %v", err, err)
	}
	if !strings.Contains(doomErr.Message, "return outside function") {
		t.Errorf("got %q", doomErr.Message)
	}
}

func TestQuestionSemicolonInsertion(t *testing.T) {
	out, _, err := evalSource(t, "fn get() { ok(42) }\nfn use() {\n  let v = get()?\n  ok(v + 1)\n}\nmatch use() {\n  ok(v) => speak v,\n  err(e) => speak e,\n}\n")
	if err != nil {
		t.Fatal(err)
	}
	if out != "43\n" {
		t.Errorf("got %q, want %q", out, "43\n")
	}
}

func TestDecreeSemicolonInsertion(t *testing.T) {
	// decree should work without explicit ; after a trigger-ending line
	out, _, err := evalSource(t, "let x = 1\ndecree \"zero_indexed\"\nlet xs = [10, 20, 30]\nspeak xs[0]\n")
	if err != nil {
		t.Fatal(err)
	}
	if out != "10\n" {
		t.Errorf("got %q, want %q", out, "10\n")
	}
}

func TestMatchTypedNil(t *testing.T) {
	out, _, err := evalSource(t, `
match nil {
  n: nil => speak "got nil",
  _ => speak "other",
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "got nil\n" {
		t.Errorf("got %q, want %q", out, "got nil\n")
	}
}

func TestMatchTypedOk(t *testing.T) {
	out, _, err := evalSource(t, `
match ok(42) {
  v: ok => speak "ok value",
  v: err => speak "err value",
  _ => speak "other",
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok value\n" {
		t.Errorf("got %q, want %q", out, "ok value\n")
	}
}

func TestMatchTypedResult(t *testing.T) {
	out, _, err := evalSource(t, `
match err("bad") {
  v: result => speak "is a result",
  _ => speak "not a result",
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "is a result\n" {
		t.Errorf("got %q, want %q", out, "is a result\n")
	}
}

func TestMapIntKeys(t *testing.T) {
	out, _, err := evalSource(t, `
decree "deterministic_hashing"
let m = { 1: "one", 2: "two" }
speak m[1]
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "one\n" {
		t.Errorf("got %q, want %q", out, "one\n")
	}
}

func TestAmbitiousModeIndexAssign(t *testing.T) {
	out, _, err := evalSource(t, `
decree "ambitious_mode"
decree "zero_indexed"
let xs = [1, 2, 3]
xs[0] == 99
speak xs[0]
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "99\n" {
		t.Errorf("got %q, want %q", out, "99\n")
	}
}

func TestAmbitiousModeDotAssign(t *testing.T) {
	out, _, err := evalSource(t, `
decree "ambitious_mode"
decree "deterministic_hashing"
let m = { "x": 1 }
m.x == 42
speak m.x
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "42\n" {
		t.Errorf("got %q, want %q", out, "42\n")
	}
}

func testExampleFile(t *testing.T, filename string) {
	t.Helper()

	// Find the examples directory relative to this test file.
	// Walk up from internal/eval to repo root.
	repoRoot := filepath.Join("..", "..")
	path := filepath.Join(repoRoot, "examples", filename)

	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}

	// Read golden file
	goldenName := strings.TrimSuffix(filename, ".mor") + ".golden"
	goldenPath := filepath.Join(repoRoot, "testdata", goldenName)
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("cannot read golden %s: %v", goldenPath, err)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	prog := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors in %s: %s", filename, strings.Join(errs, "; "))
	}

	var buf bytes.Buffer
	ev := New()
	ev.SetOutput(&buf)
	_, evalErr := ev.Eval(prog)
	if evalErr != nil {
		t.Fatalf("eval error in %s: %v", filename, evalErr)
	}

	got := buf.String()
	if got != string(expected) {
		t.Errorf("%s: got %q, want %q", filename, got, string(expected))
	}
}

func TestExternFnStub(t *testing.T) {
	out, _, err := evalSource(t, `
extern fn do_thing(x);
speak do_thing(42);
`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "nil\n" {
		t.Errorf("got %q, want %q", out, "nil\n")
	}
}

func TestSoftCastsAllPaths(t *testing.T) {
	out, _, err := evalSource(t, `
decree "soft_casts"
let x = nil as int
speak x
`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "err(") {
		t.Errorf("expected err result from soft cast, got %q", out)
	}
}

func TestLenUnicode(t *testing.T) {
	out, _, err := evalSource(t, `speak len("héllo");`)
	if err != nil {
		t.Fatal(err)
	}
	if out != "5\n" {
		t.Errorf("got %q, want %q", out, "5\n")
	}
}

func TestMatchExhaustion(t *testing.T) {
	_, _, err := evalSource(t, `
match 99 {
  1 => speak "one",
  2 => speak "two",
}
`)
	if err == nil {
		t.Fatal("expected doom from exhausted match")
	}
}

// --- Anonymous functions ---

func TestAnonymousFnBasic(t *testing.T) {
	out, _, err := evalSource(t, `
let double = fn(x) { x * 2 }
speak double(5)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "10\n" {
		t.Errorf("got %q, want %q", out, "10\n")
	}
}

func TestAnonymousFnNoParams(t *testing.T) {
	out, _, err := evalSource(t, `
let greet = fn() { "hello" }
speak greet()
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Errorf("got %q, want %q", out, "hello\n")
	}
}

func TestAnonymousFnMultipleParams(t *testing.T) {
	out, _, err := evalSource(t, `
let add = fn(a, b) { a + b }
speak add(3, 7)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "10\n" {
		t.Errorf("got %q, want %q", out, "10\n")
	}
}

// --- Closures ---

func TestClosure(t *testing.T) {
	out, _, err := evalSource(t, `
fn make_adder(n) {
  fn(x) { x + n }
}
let add5 = make_adder(5)
speak add5(10)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "15\n" {
		t.Errorf("got %q, want %q", out, "15\n")
	}
}

func TestClosureOverMutableVar(t *testing.T) {
	out, _, err := evalSource(t, `
let count = 0
let inc = fn() {
  count = count + 1
  count
}
inc()
inc()
speak inc()
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "3\n" {
		t.Errorf("got %q, want %q", out, "3\n")
	}
}

// --- Higher-order functions ---

func TestHigherOrderFunction(t *testing.T) {
	out, _, err := evalSource(t, `
fn apply(f, x) {
  f(x)
}
let square = fn(n) { n * n }
speak apply(square, 4)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "16\n" {
		t.Errorf("got %q, want %q", out, "16\n")
	}
}

func TestFnReturningFn(t *testing.T) {
	out, _, err := evalSource(t, `
fn multiplier(factor) {
  fn(x) { x * factor }
}
let triple = multiplier(3)
speak triple(7)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "21\n" {
		t.Errorf("got %q, want %q", out, "21\n")
	}
}

// --- valuesEqual for ok/err ---

func TestValuesEqualOkErr(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{`speak ok(1) == ok(1)`, "true\n"},
		{`speak ok(1) == ok(2)`, "false\n"},
		{`speak err("x") == err("x")`, "true\n"},
		{`speak err("x") == err("y")`, "false\n"},
		{`speak ok(1) == err(1)`, "false\n"},
		{`speak ok(1) != ok(2)`, "true\n"},
	}
	for _, tt := range tests {
		out, _, err := evalSource(t, tt.source)
		if err != nil {
			t.Errorf("source %q: unexpected error: %v", tt.source, err)
			continue
		}
		if out != tt.want {
			t.Errorf("source %q: got %q, want %q", tt.source, out, tt.want)
		}
	}
}

// --- if without else ---

func TestIfWithoutElse(t *testing.T) {
	out, _, err := evalSource(t, `
if true {
  speak "yes"
}
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "yes\n" {
		t.Errorf("got %q, want %q", out, "yes\n")
	}
}

func TestIfWithoutElseFalse(t *testing.T) {
	out, _, err := evalSource(t, `
if false {
  speak "nope"
}
speak "done"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "done\n" {
		t.Errorf("got %q, want %q", out, "done\n")
	}
}

// --- Immediate fn call ---

func TestImmediateFnCall(t *testing.T) {
	out, _, err := evalSource(t, `speak fn(x) { x + 1 }(41)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "42\n" {
		t.Errorf("got %q, want %q", out, "42\n")
	}
}
