package eval

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joeabbey/morgoth/internal/parser"
)

// Control flow signals implemented as error types.

// DoomError is a non-local exit (like an exception).
type DoomError struct {
	Message string
}

func (e *DoomError) Error() string { return "doom: " + e.Message }

// ReturnSignal carries a return value out of a function body.
type ReturnSignal struct {
	Value *Value
}

func (e *ReturnSignal) Error() string { return "return signal" }

// PropagateError carries an error value from the ? operator.
type PropagateError struct {
	Value *Value
}

func (e *PropagateError) Error() string { return "propagate error" }

// GuardReturnSignal carries a value from a failed guard out of the enclosing function.
type GuardReturnSignal struct {
	Value *Value
}

func (e *GuardReturnSignal) Error() string { return "guard return" }

// Evaluator walks the AST and produces values.
type Evaluator struct {
	env     *Env
	decrees *DecreeConfig
	output  io.Writer
}

// New creates a new Evaluator with default settings.
func New() *Evaluator {
	return &Evaluator{
		env:     NewEnv(nil),
		decrees: NewDecreeConfig(),
		output:  os.Stdout,
	}
}

// SetOutput sets the writer for speak output (useful for testing).
func (ev *Evaluator) SetOutput(w io.Writer) {
	ev.output = w
}

// Eval evaluates a complete program.
func (ev *Evaluator) Eval(program *parser.Program) (*Value, error) {
	var result *Value
	for _, item := range program.Items {
		val, err := ev.evalItem(item)
		if err != nil {
			if gs, ok := err.(*GuardReturnSignal); ok {
				return nil, &DoomError{Message: fmt.Sprintf("unhandled guard return: %s", gs.Value.String())}
			}
			if pe, ok := err.(*PropagateError); ok {
				return nil, &DoomError{Message: fmt.Sprintf("unhandled error propagation: %s", pe.Value.String())}
			}
			if rs, ok := err.(*ReturnSignal); ok {
				_ = rs
				return nil, &DoomError{Message: "return outside function"}
			}
			return nil, err
		}
		result = val
	}
	if result == nil {
		return NilVal(), nil
	}
	return result, nil
}

func (ev *Evaluator) evalItem(item parser.Item) (*Value, error) {
	switch n := item.(type) {
	case *parser.FnDecl:
		return ev.evalFnDecl(n)
	case *parser.ExternDecl:
		// Register a stub function that returns nil for all extern declarations.
		params := make([]string, len(n.Params))
		for i, p := range n.Params {
			params[i] = p.Name
		}
		stub := &FnValue{
			Name:   n.Name,
			Params: params,
			Body:   nil, // no body — callFunction handles nil body
			Env:    ev.env,
		}
		ev.env.Define(n.Name, FnVal(stub), false)
		return NilVal(), nil
	case *parser.LetStmt:
		return ev.evalLetStmt(n)
	case *parser.ConstStmt:
		return ev.evalConstStmt(n)
	case *parser.ReturnStmt:
		return ev.evalReturnStmt(n)
	case *parser.DecreeStmt:
		return ev.evalDecreeStmt(n)
	case *parser.ExprStmt:
		return ev.evalExpr(n.Expression)
	default:
		return nil, &DoomError{Message: fmt.Sprintf("unknown item type: %T", item)}
	}
}

// --- Statement evaluation ---

func (ev *Evaluator) evalFnDecl(decl *parser.FnDecl) (*Value, error) {
	params := make([]string, len(decl.Params))
	for i, p := range decl.Params {
		params[i] = p.Name
	}
	fn := &FnValue{
		Name:   decl.Name,
		Params: params,
		Body:   decl.Body,
		Env:    ev.env,
	}
	ev.env.Define(decl.Name, FnVal(fn), false)
	return NilVal(), nil
}

func (ev *Evaluator) evalLetStmt(stmt *parser.LetStmt) (*Value, error) {
	val, err := ev.evalExpr(stmt.Value)
	if err != nil {
		return nil, err
	}
	ev.env.Define(stmt.Name, val, false)
	return NilVal(), nil
}

func (ev *Evaluator) evalConstStmt(stmt *parser.ConstStmt) (*Value, error) {
	val, err := ev.evalExpr(stmt.Value)
	if err != nil {
		return nil, err
	}
	ev.env.Define(stmt.Name, val, true)
	return NilVal(), nil
}

func (ev *Evaluator) evalReturnStmt(stmt *parser.ReturnStmt) (*Value, error) {
	val, err := ev.evalExpr(stmt.Value)
	if err != nil {
		return nil, err
	}
	return nil, &ReturnSignal{Value: val}
}

func (ev *Evaluator) evalDecreeStmt(stmt *parser.DecreeStmt) (*Value, error) {
	ev.decrees.Apply(stmt.Value)
	return NilVal(), nil
}

func (ev *Evaluator) evalStmt(stmt parser.Stmt) (*Value, error) {
	switch n := stmt.(type) {
	case *parser.LetStmt:
		return ev.evalLetStmt(n)
	case *parser.ConstStmt:
		return ev.evalConstStmt(n)
	case *parser.ReturnStmt:
		return ev.evalReturnStmt(n)
	case *parser.DecreeStmt:
		return ev.evalDecreeStmt(n)
	case *parser.ExprStmt:
		return ev.evalExpr(n.Expression)
	default:
		return nil, &DoomError{Message: fmt.Sprintf("unknown stmt type: %T", stmt)}
	}
}

// --- Expression evaluation ---

func (ev *Evaluator) evalExpr(expr parser.Expr) (*Value, error) {
	if expr == nil {
		return NilVal(), nil
	}
	switch n := expr.(type) {
	case *parser.IntLitExpr:
		return IntVal(n.Value), nil
	case *parser.FloatLitExpr:
		return FloatVal(n.Value), nil
	case *parser.StringLitExpr:
		return StrVal(n.Value), nil
	case *parser.BoolLitExpr:
		return BoolVal(n.Value), nil
	case *parser.NilLitExpr:
		return NilVal(), nil
	case *parser.IdentExpr:
		return ev.evalIdentExpr(n)
	case *parser.ArrayLitExpr:
		return ev.evalArrayLitExpr(n)
	case *parser.MapLitExpr:
		return ev.evalMapLitExpr(n)
	case *parser.BinaryExpr:
		return ev.evalBinaryExpr(n)
	case *parser.UnaryExpr:
		return ev.evalUnaryExpr(n)
	case *parser.AssignExpr:
		return ev.evalAssignExpr(n)
	case *parser.IndexAssignExpr:
		return ev.evalIndexAssignExpr(n)
	case *parser.DotAssignExpr:
		return ev.evalDotAssignExpr(n)
	case *parser.CallExpr:
		return ev.evalCallExpr(n)
	case *parser.IndexExpr:
		return ev.evalIndexExpr(n)
	case *parser.DotExpr:
		return ev.evalDotExpr(n)
	case *parser.PropagateExpr:
		return ev.evalPropagateExpr(n)
	case *parser.IfExpr:
		return ev.evalIfExpr(n)
	case *parser.MatchExpr:
		return ev.evalMatchExpr(n)
	case *parser.GuardExpr:
		return ev.evalGuardExpr(n)
	case *parser.BlockExpr:
		return ev.evalBlockExpr(n)
	case *parser.OkExpr:
		return ev.evalOkExpr(n)
	case *parser.ErrExpr:
		return ev.evalErrExpr(n)
	case *parser.AsExpr:
		return ev.evalAsExpr(n)
	case *parser.SpeakExpr:
		return ev.evalSpeakExpr(n)
	case *parser.DoomExpr:
		return ev.evalDoomExpr(n)
	case *parser.SorryExpr:
		return ev.evalSorryExpr(n)
	case *parser.ChantExpr:
		return ev.evalChantExpr(n)
	case *parser.SpawnExpr:
		// MVP stub: run spawn body synchronously, return nil.
		_, err := ev.evalBlockExpr(n.Body)
		if err != nil {
			return nil, err
		}
		return NilVal(), nil
	case *parser.AwaitAllExpr:
		// MVP stub: no-op since spawn runs synchronously.
		return NilVal(), nil
	default:
		return nil, &DoomError{Message: fmt.Sprintf("unknown expr type: %T", expr)}
	}
}

func (ev *Evaluator) evalIdentExpr(expr *parser.IdentExpr) (*Value, error) {
	val, err := ev.env.Get(expr.Name)
	if err != nil {
		return nil, &DoomError{Message: err.Error()}
	}
	return val, nil
}

func (ev *Evaluator) evalArrayLitExpr(expr *parser.ArrayLitExpr) (*Value, error) {
	elems := make([]*Value, len(expr.Elements))
	for i, e := range expr.Elements {
		val, err := ev.evalExpr(e)
		if err != nil {
			return nil, err
		}
		elems[i] = val
	}
	return ArrayVal(elems), nil
}

func (ev *Evaluator) evalMapLitExpr(expr *parser.MapLitExpr) (*Value, error) {
	m := NewOrderedMap()
	for _, pair := range expr.Pairs {
		key, err := ev.evalExpr(pair.Key)
		if err != nil {
			return nil, err
		}
		val, err := ev.evalExpr(pair.Value)
		if err != nil {
			return nil, err
		}
		m.Set(key.String(), val)
	}
	return MapVal(m), nil
}

func (ev *Evaluator) evalBinaryExpr(expr *parser.BinaryExpr) (*Value, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	// Short-circuit for logical operators
	if expr.Op == "and" {
		if !left.IsTruthy() {
			return left, nil
		}
		return ev.evalExpr(expr.Right)
	}
	if expr.Op == "or" {
		if left.IsTruthy() {
			return left, nil
		}
		return ev.evalExpr(expr.Right)
	}

	right, err := ev.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case "+":
		return ev.evalAdd(left, right)
	case "-":
		return ev.evalArith(left, right, "-")
	case "*":
		return ev.evalArith(left, right, "*")
	case "/":
		return ev.evalArith(left, right, "/")
	case "%":
		return ev.evalArith(left, right, "%")
	case "==":
		if ev.decrees.AmbitiousMode && right.IsTruthy() {
			switch lhs := expr.Left.(type) {
			case *parser.IdentExpr:
				if err := ev.env.Set(lhs.Name, right); err != nil {
					return nil, &DoomError{Message: err.Error()}
				}
				return right, nil
			case *parser.IndexExpr:
				collection, err := ev.evalExpr(lhs.Left)
				if err != nil {
					return nil, err
				}
				index, err := ev.evalExpr(lhs.Index)
				if err != nil {
					return nil, err
				}
				switch collection.Kind {
				case ValArray:
					if index.Kind == ValInt {
						idx := ev.adjustIndex(index.Int)
						if idx >= 0 && idx < int64(len(collection.Array)) {
							collection.Array[idx] = right
						}
					}
				case ValMap:
					collection.Map.Set(index.String(), right)
				}
				return right, nil
			case *parser.DotExpr:
				obj, err := ev.evalExpr(lhs.Left)
				if err != nil {
					return nil, err
				}
				if obj.Kind == ValMap {
					obj.Map.Set(lhs.Field, right)
				}
				return right, nil
			}
		}
		return BoolVal(ev.valuesEqual(left, right)), nil
	case "===":
		return BoolVal(ev.valuesStrictEqual(left, right)), nil
	case "!=":
		return BoolVal(!ev.valuesEqual(left, right)), nil
	case "<":
		return ev.evalCompare(left, right, "<")
	case ">":
		return ev.evalCompare(left, right, ">")
	case "<=":
		return ev.evalCompare(left, right, "<=")
	case ">=":
		return ev.evalCompare(left, right, ">=")
	default:
		return nil, &DoomError{Message: fmt.Sprintf("unknown operator: %s", expr.Op)}
	}
}

func (ev *Evaluator) evalAdd(left, right *Value) (*Value, error) {
	if left.Kind == ValStr || right.Kind == ValStr {
		return StrVal(left.String() + right.String()), nil
	}
	if left.Kind == ValFloat || right.Kind == ValFloat {
		lf := toFloat(left)
		rf := toFloat(right)
		return FloatVal(lf + rf), nil
	}
	if left.Kind == ValInt && right.Kind == ValInt {
		return IntVal(left.Int + right.Int), nil
	}
	return nil, &DoomError{Message: fmt.Sprintf("cannot add %v and %v", left.Kind, right.Kind)}
}

func (ev *Evaluator) evalArith(left, right *Value, op string) (*Value, error) {
	if left.Kind == ValFloat || right.Kind == ValFloat {
		lf := toFloat(left)
		rf := toFloat(right)
		switch op {
		case "-":
			return FloatVal(lf - rf), nil
		case "*":
			return FloatVal(lf * rf), nil
		case "/":
			if rf == 0 {
				return nil, &DoomError{Message: "division by zero"}
			}
			return FloatVal(lf / rf), nil
		case "%":
			return nil, &DoomError{Message: "modulo on floats not supported"}
		}
	}
	if left.Kind == ValInt && right.Kind == ValInt {
		switch op {
		case "-":
			return IntVal(left.Int - right.Int), nil
		case "*":
			return IntVal(left.Int * right.Int), nil
		case "/":
			if right.Int == 0 {
				return nil, &DoomError{Message: "division by zero"}
			}
			return IntVal(left.Int / right.Int), nil
		case "%":
			if right.Int == 0 {
				return nil, &DoomError{Message: "division by zero"}
			}
			return IntVal(left.Int % right.Int), nil
		}
	}
	return nil, &DoomError{Message: fmt.Sprintf("cannot perform %s on %v and %v", op, left.Kind, right.Kind)}
}

func (ev *Evaluator) evalCompare(left, right *Value, op string) (*Value, error) {
	if left.Kind == ValInt && right.Kind == ValInt {
		switch op {
		case "<":
			return BoolVal(left.Int < right.Int), nil
		case ">":
			return BoolVal(left.Int > right.Int), nil
		case "<=":
			return BoolVal(left.Int <= right.Int), nil
		case ">=":
			return BoolVal(left.Int >= right.Int), nil
		}
	}
	if left.Kind == ValFloat || right.Kind == ValFloat {
		lf := toFloat(left)
		rf := toFloat(right)
		switch op {
		case "<":
			return BoolVal(lf < rf), nil
		case ">":
			return BoolVal(lf > rf), nil
		case "<=":
			return BoolVal(lf <= rf), nil
		case ">=":
			return BoolVal(lf >= rf), nil
		}
	}
	if left.Kind == ValStr && right.Kind == ValStr {
		switch op {
		case "<":
			return BoolVal(left.Str < right.Str), nil
		case ">":
			return BoolVal(left.Str > right.Str), nil
		case "<=":
			return BoolVal(left.Str <= right.Str), nil
		case ">=":
			return BoolVal(left.Str >= right.Str), nil
		}
	}
	return nil, &DoomError{Message: fmt.Sprintf("cannot compare %v and %v", left.Kind, right.Kind)}
}

func (ev *Evaluator) valuesEqual(a, b *Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ValInt:
		return a.Int == b.Int
	case ValFloat:
		return a.Float == b.Float
	case ValBool:
		return a.Bool == b.Bool
	case ValStr:
		return a.Str == b.Str
	case ValNil:
		return true
	default:
		return false
	}
}

func (ev *Evaluator) valuesStrictEqual(a, b *Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case ValInt:
		return a.Int == b.Int
	case ValFloat:
		return a.Float == b.Float
	case ValBool:
		return a.Bool == b.Bool
	case ValStr:
		return a.Str == b.Str
	case ValNil:
		return true
	case ValOk:
		return ev.valuesStrictEqual(a.Inner, b.Inner)
	case ValErr:
		return ev.valuesStrictEqual(a.Inner, b.Inner)
	case ValPtr:
		return a.Int == b.Int
	default:
		// Arrays, Maps, Fns: reference identity (always false for distinct values)
		return a == b
	}
}

func toFloat(v *Value) float64 {
	switch v.Kind {
	case ValFloat:
		return v.Float
	case ValInt:
		return float64(v.Int)
	default:
		return 0
	}
}

func (ev *Evaluator) evalUnaryExpr(expr *parser.UnaryExpr) (*Value, error) {
	right, err := ev.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Op {
	case "-":
		switch right.Kind {
		case ValInt:
			return IntVal(-right.Int), nil
		case ValFloat:
			return FloatVal(-right.Float), nil
		default:
			return nil, &DoomError{Message: "cannot negate non-numeric value"}
		}
	case "!":
		return BoolVal(!right.IsTruthy()), nil
	case "&":
		// Address-of operator: for MVP, return a ptr(0)
		return PtrVal(0), nil
	default:
		return nil, &DoomError{Message: fmt.Sprintf("unknown unary operator: %s", expr.Op)}
	}
}

func (ev *Evaluator) evalAssignExpr(expr *parser.AssignExpr) (*Value, error) {
	val, err := ev.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	if err := ev.env.Set(expr.Name, val); err != nil {
		return nil, &DoomError{Message: err.Error()}
	}
	return val, nil
}

func (ev *Evaluator) evalIndexAssignExpr(expr *parser.IndexAssignExpr) (*Value, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	index, err := ev.evalExpr(expr.Index)
	if err != nil {
		return nil, err
	}
	val, err := ev.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	switch left.Kind {
	case ValArray:
		if index.Kind != ValInt {
			return nil, &DoomError{Message: "array index must be int"}
		}
		idx := ev.adjustIndex(index.Int)
		if idx < 0 || idx >= int64(len(left.Array)) {
			return nil, &DoomError{Message: fmt.Sprintf("array index out of bounds: %d", idx)}
		}
		left.Array[idx] = val
		return val, nil
	case ValMap:
		key := index.String()
		left.Map.Set(key, val)
		return val, nil
	default:
		return nil, &DoomError{Message: fmt.Sprintf("cannot assign to index of %s", left.String())}
	}
}

func (ev *Evaluator) evalDotAssignExpr(expr *parser.DotAssignExpr) (*Value, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	val, err := ev.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}

	if left.Kind != ValMap {
		return nil, &DoomError{Message: fmt.Sprintf("cannot assign field %s on %s", expr.Field, left.String())}
	}
	left.Map.Set(expr.Field, val)
	return val, nil
}

func (ev *Evaluator) evalCallExpr(expr *parser.CallExpr) (*Value, error) {
	// Evaluate args first (needed for both builtins and user functions).
	args := make([]*Value, len(expr.Args))
	for i, a := range expr.Args {
		val, err := ev.evalExpr(a)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Check for built-in functions by name before evaluating the function expression,
	// since builtins like len() are not defined in the environment.
	if ident, ok := expr.Function.(*parser.IdentExpr); ok {
		result, isBuiltin, err := ev.callBuiltin(ident.Name, args)
		if isBuiltin {
			return result, err
		}
	}

	fn, err := ev.evalExpr(expr.Function)
	if err != nil {
		return nil, err
	}

	if fn.Kind != ValFn {
		return nil, &DoomError{Message: fmt.Sprintf("cannot call non-function: %s", fn.String())}
	}

	return ev.callFunction(fn.Fn, args)
}

func (ev *Evaluator) callFunction(fn *FnValue, args []*Value) (*Value, error) {
	// Extern stub: no body, just return nil.
	if fn.Body == nil {
		return NilVal(), nil
	}

	callEnv := NewEnv(fn.Env)
	for i, param := range fn.Params {
		if i < len(args) {
			callEnv.Define(param, args[i], false)
		} else {
			callEnv.Define(param, NilVal(), false)
		}
	}

	savedEnv := ev.env
	ev.env = callEnv
	result, err := ev.evalBlockExpr(fn.Body)
	ev.env = savedEnv

	if err != nil {
		switch e := err.(type) {
		case *ReturnSignal:
			return e.Value, nil
		case *GuardReturnSignal:
			return e.Value, nil
		case *PropagateError:
			return ErrVal(e.Value), nil
		case *DoomError:
			return nil, err
		default:
			return nil, err
		}
	}
	return result, nil
}

func (ev *Evaluator) evalIndexExpr(expr *parser.IndexExpr) (*Value, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	index, err := ev.evalExpr(expr.Index)
	if err != nil {
		return nil, err
	}

	switch left.Kind {
	case ValArray:
		if index.Kind != ValInt {
			return nil, &DoomError{Message: "array index must be int"}
		}
		idx := ev.adjustIndex(index.Int)
		if idx < 0 || idx >= int64(len(left.Array)) {
			return nil, &DoomError{Message: fmt.Sprintf("array index out of bounds: %d", idx)}
		}
		return left.Array[idx], nil
	case ValMap:
		key := index.String()
		val, ok := left.Map.Get(key)
		if !ok {
			return NilVal(), nil
		}
		return val, nil
	case ValStr:
		if index.Kind != ValInt {
			return nil, &DoomError{Message: "string index must be int"}
		}
		runes := []rune(left.Str)
		idx := ev.adjustIndex(index.Int)
		if idx < 0 || idx >= int64(len(runes)) {
			return nil, &DoomError{Message: fmt.Sprintf("string index out of bounds: %d", idx)}
		}
		return StrVal(string(runes[idx])), nil
	default:
		return nil, &DoomError{Message: fmt.Sprintf("cannot index into %s", left.String())}
	}
}

func (ev *Evaluator) adjustIndex(idx int64) int64 {
	switch ev.decrees.IndexingBase {
	case "zero":
		return idx
	case "one":
		return idx - 1
	case "weekday":
		weekday := time.Now().Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			return idx // 0-based on weekends
		}
		return idx - 1 // 1-based on weekdays
	default:
		return idx
	}
}

func (ev *Evaluator) evalDotExpr(expr *parser.DotExpr) (*Value, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	if left.Kind == ValMap {
		val, ok := left.Map.Get(expr.Field)
		if !ok {
			return NilVal(), nil
		}
		return val, nil
	}
	return nil, &DoomError{Message: fmt.Sprintf("cannot access field %s on %s", expr.Field, left.String())}
}

func (ev *Evaluator) evalPropagateExpr(expr *parser.PropagateExpr) (*Value, error) {
	inner, err := ev.evalExpr(expr.Inner)
	if err != nil {
		return nil, err
	}

	switch inner.Kind {
	case ValOk:
		return inner.Inner, nil
	case ValErr:
		return nil, &PropagateError{Value: inner.Inner}
	case ValNil:
		return nil, &PropagateError{Value: StrVal("nil")}
	default:
		return inner, nil
	}
}

func (ev *Evaluator) evalIfExpr(expr *parser.IfExpr) (*Value, error) {
	cond, err := ev.evalExpr(expr.Condition)
	if err != nil {
		return nil, err
	}
	if cond.IsTruthy() {
		return ev.evalBlockExpr(expr.Then)
	}
	if expr.Else != nil {
		switch e := expr.Else.(type) {
		case *parser.BlockExpr:
			return ev.evalBlockExpr(e)
		case *parser.IfExpr:
			return ev.evalIfExpr(e)
		default:
			return ev.evalExpr(expr.Else)
		}
	}
	return NilVal(), nil
}

func (ev *Evaluator) evalMatchExpr(expr *parser.MatchExpr) (*Value, error) {
	subject, err := ev.evalExpr(expr.Subject)
	if err != nil {
		return nil, err
	}

	for _, arm := range expr.Arms {
		matched, bindings := ev.matchPattern(arm.Pattern, subject)
		if matched {
			matchEnv := NewEnv(ev.env)
			for name, val := range bindings {
				matchEnv.Define(name, val, false)
			}
			savedEnv := ev.env
			ev.env = matchEnv
			result, err := ev.evalExpr(arm.Body)
			ev.env = savedEnv
			return result, err
		}
	}
	return nil, &DoomError{Message: fmt.Sprintf("match exhausted: no arm matched value %s", subject.String())}
}

func (ev *Evaluator) matchPattern(pat parser.Pattern, subject *Value) (bool, map[string]*Value) {
	bindings := make(map[string]*Value)

	switch p := pat.(type) {
	case *parser.WildcardPattern:
		return true, bindings

	case *parser.LiteralPattern:
		litVal, err := ev.evalExpr(p.Value)
		if err != nil {
			return false, nil
		}
		return ev.valuesEqual(subject, litVal), bindings

	case *parser.IdentPattern:
		// Check for ok(v) / err(e) patterns
		if strings.HasPrefix(p.Name, "ok(") && strings.HasSuffix(p.Name, ")") {
			inner := p.Name[3 : len(p.Name)-1]
			if subject.Kind != ValOk {
				return false, nil
			}
			if inner != "" {
				bindings[inner] = subject.Inner
			}
			return true, bindings
		}
		if strings.HasPrefix(p.Name, "err(") && strings.HasSuffix(p.Name, ")") {
			inner := p.Name[4 : len(p.Name)-1]
			if subject.Kind != ValErr {
				return false, nil
			}
			if inner != "" {
				bindings[inner] = subject.Inner
			}
			return true, bindings
		}
		bindings[p.Name] = subject
		return true, bindings

	case *parser.TypedPattern:
		if ev.matchesType(subject, p.TypeName) {
			bindings[p.Name] = subject
			return true, bindings
		}
		return false, nil

	case *parser.GuardedPattern:
		matched, innerBindings := ev.matchPattern(p.Inner, subject)
		if !matched {
			return false, nil
		}
		// Evaluate guard with bindings in scope
		guardEnv := NewEnv(ev.env)
		for name, val := range innerBindings {
			guardEnv.Define(name, val, false)
		}
		savedEnv := ev.env
		ev.env = guardEnv
		guardVal, err := ev.evalExpr(p.Guard)
		ev.env = savedEnv
		if err != nil || !guardVal.IsTruthy() {
			return false, nil
		}
		return true, innerBindings

	default:
		return false, nil
	}
}

func (ev *Evaluator) matchesType(val *Value, typeName string) bool {
	switch typeName {
	case "int":
		return val.Kind == ValInt
	case "float":
		return val.Kind == ValFloat
	case "bool":
		return val.Kind == ValBool
	case "str", "string":
		return val.Kind == ValStr
	case "array":
		return val.Kind == ValArray
	case "map":
		return val.Kind == ValMap
	case "fn":
		return val.Kind == ValFn
	case "ptr":
		return val.Kind == ValPtr
	case "nil":
		return val.Kind == ValNil
	case "ok":
		return val.Kind == ValOk
	case "err":
		return val.Kind == ValErr
	case "result":
		return val.Kind == ValOk || val.Kind == ValErr
	default:
		return false
	}
}

func (ev *Evaluator) evalGuardExpr(expr *parser.GuardExpr) (*Value, error) {
	cond, err := ev.evalExpr(expr.Condition)
	if err != nil {
		return nil, err
	}
	if !cond.IsTruthy() {
		val, err := ev.evalExpr(expr.ElseBody)
		if err != nil {
			return nil, err
		}
		// Guard semantics: non-local return from enclosing function with else value.
		return nil, &GuardReturnSignal{Value: val}
	}
	return NilVal(), nil
}

func (ev *Evaluator) evalBlockExpr(block *parser.BlockExpr) (*Value, error) {
	blockEnv := NewEnv(ev.env)
	savedEnv := ev.env
	ev.env = blockEnv

	for _, stmt := range block.Stmts {
		_, err := ev.evalStmt(stmt)
		if err != nil {
			ev.env = savedEnv
			return nil, err
		}
	}

	var result *Value
	if block.FinalExpr != nil {
		var err error
		result, err = ev.evalExpr(block.FinalExpr)
		if err != nil {
			ev.env = savedEnv
			return nil, err
		}
	} else {
		result = NilVal()
	}

	ev.env = savedEnv
	return result, nil
}

func (ev *Evaluator) evalOkExpr(expr *parser.OkExpr) (*Value, error) {
	inner, err := ev.evalExpr(expr.Inner)
	if err != nil {
		return nil, err
	}
	return OkVal(inner), nil
}

func (ev *Evaluator) evalErrExpr(expr *parser.ErrExpr) (*Value, error) {
	inner, err := ev.evalExpr(expr.Inner)
	if err != nil {
		return nil, err
	}
	return ErrVal(inner), nil
}

func (ev *Evaluator) evalAsExpr(expr *parser.AsExpr) (*Value, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}

	switch expr.TypeName {
	case "int":
		switch left.Kind {
		case ValInt:
			return left, nil
		case ValFloat:
			return IntVal(int64(left.Float)), nil
		case ValStr:
			n, err := strconv.ParseInt(strings.TrimSpace(left.Str), 10, 64)
			if err != nil {
				if ev.decrees.SoftCasts {
					return ErrVal(StrVal(fmt.Sprintf("cannot convert %q to int", left.Str))), nil
				}
				return nil, &DoomError{Message: fmt.Sprintf("cannot convert %q to int", left.Str)}
			}
			return IntVal(n), nil
		case ValBool:
			if left.Bool {
				return IntVal(1), nil
			}
			return IntVal(0), nil
		default:
			msg := fmt.Sprintf("cannot cast %s to int", left.String())
			if ev.decrees.SoftCasts {
				return ErrVal(StrVal(msg)), nil
			}
			return nil, &DoomError{Message: msg}
		}
	case "float":
		switch left.Kind {
		case ValFloat:
			return left, nil
		case ValInt:
			return FloatVal(float64(left.Int)), nil
		case ValStr:
			f, err := strconv.ParseFloat(strings.TrimSpace(left.Str), 64)
			if err != nil {
				msg := fmt.Sprintf("cannot convert %q to float", left.Str)
				if ev.decrees.SoftCasts {
					return ErrVal(StrVal(msg)), nil
				}
				return nil, &DoomError{Message: msg}
			}
			return FloatVal(f), nil
		default:
			msg := fmt.Sprintf("cannot cast %s to float", left.String())
			if ev.decrees.SoftCasts {
				return ErrVal(StrVal(msg)), nil
			}
			return nil, &DoomError{Message: msg}
		}
	case "str", "string":
		return StrVal(left.String()), nil
	case "bool":
		return BoolVal(left.IsTruthy()), nil
	default:
		msg := fmt.Sprintf("unknown cast target: %s", expr.TypeName)
		if ev.decrees.SoftCasts {
			return ErrVal(StrVal(msg)), nil
		}
		return nil, &DoomError{Message: msg}
	}
}

func (ev *Evaluator) evalSpeakExpr(expr *parser.SpeakExpr) (*Value, error) {
	val, err := ev.evalExpr(expr.Value)
	if err != nil {
		return nil, err
	}
	_, writeErr := fmt.Fprintln(ev.output, val.String())
	if writeErr != nil {
		if expr.ElseBody != nil {
			return ev.evalExpr(expr.ElseBody)
		}
		return ErrVal(StrVal(writeErr.Error())), nil
	}
	return OkVal(NilVal()), nil
}

func (ev *Evaluator) evalDoomExpr(expr *parser.DoomExpr) (*Value, error) {
	msg, err := ev.evalExpr(expr.Message)
	if err != nil {
		return nil, err
	}
	return nil, &DoomError{Message: msg.String()}
}

func (ev *Evaluator) evalSorryExpr(expr *parser.SorryExpr) (*Value, error) {
	if ev.decrees.NoForgiveness {
		return ErrVal(StrVal("no")), nil
	}
	if err := ev.env.Forgive(expr.Name); err != nil {
		return ErrVal(StrVal(err.Error())), nil
	}
	return OkVal(NilVal()), nil
}

func (ev *Evaluator) evalChantExpr(expr *parser.ChantExpr) (*Value, error) {
	_, err := ev.evalExpr(expr.Name)
	if err != nil {
		return nil, err
	}
	return OkVal(NilVal()), nil
}
