package eval

import (
	"fmt"
	"strings"

	"github.com/joeabbey/morgoth/internal/parser"
)

// ValueKind tags the runtime type of a Value.
type ValueKind int

const (
	ValInt ValueKind = iota
	ValFloat
	ValBool
	ValStr
	ValNil
	ValArray
	ValMap
	ValFn
	ValOk
	ValErr
	ValPtr
)

// Value is the universal runtime value.
type Value struct {
	Kind  ValueKind
	Int   int64
	Float float64
	Bool  bool
	Str   string
	Array []*Value
	Map   *OrderedMap
	Fn    *FnValue
	Inner *Value // for Ok/Err wrapping
}

// FnValue captures a function closure.
type FnValue struct {
	Name   string
	Params []string
	Body   *parser.BlockExpr
	Env    *Env
}

// OrderedMap preserves insertion order for deterministic output.
type OrderedMap struct {
	keys   []string
	values map[string]*Value
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{values: make(map[string]*Value)}
}

func (m *OrderedMap) Set(key string, val *Value) {
	if _, exists := m.values[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.values[key] = val
}

func (m *OrderedMap) Get(key string) (*Value, bool) {
	v, ok := m.values[key]
	return v, ok
}

func (m *OrderedMap) Keys() []string {
	return m.keys
}

func (m *OrderedMap) Len() int {
	return len(m.keys)
}

// IsTruthy implements Morgoth truthiness (spec 4.2).
func (v *Value) IsTruthy() bool {
	switch v.Kind {
	case ValBool:
		return v.Bool
	case ValInt:
		return v.Int != 0
	case ValStr:
		return v.Str != ""
	case ValPtr:
		return v.Int != 0
	case ValNil:
		return false
	default:
		return true
	}
}

// String returns a human-readable representation for speak output.
func (v *Value) String() string {
	switch v.Kind {
	case ValInt:
		return fmt.Sprintf("%d", v.Int)
	case ValFloat:
		return fmt.Sprintf("%g", v.Float)
	case ValBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case ValStr:
		return v.Str
	case ValNil:
		return "nil"
	case ValArray:
		parts := make([]string, len(v.Array))
		for i, elem := range v.Array {
			parts[i] = elem.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case ValMap:
		parts := make([]string, 0, v.Map.Len())
		for _, k := range v.Map.Keys() {
			val, _ := v.Map.Get(k)
			parts = append(parts, fmt.Sprintf("%s: %s", k, val.String()))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case ValFn:
		if v.Fn.Name != "" {
			return fmt.Sprintf("<fn %s>", v.Fn.Name)
		}
		return "<fn>"
	case ValOk:
		return fmt.Sprintf("ok(%s)", v.Inner.String())
	case ValErr:
		return fmt.Sprintf("err(%s)", v.Inner.String())
	case ValPtr:
		return fmt.Sprintf("ptr(%d)", v.Int)
	default:
		return "<unknown>"
	}
}

// Convenience constructors.

func IntVal(n int64) *Value   { return &Value{Kind: ValInt, Int: n} }
func FloatVal(f float64) *Value { return &Value{Kind: ValFloat, Float: f} }
func BoolVal(b bool) *Value   { return &Value{Kind: ValBool, Bool: b} }
func StrVal(s string) *Value  { return &Value{Kind: ValStr, Str: s} }
func NilVal() *Value          { return &Value{Kind: ValNil} }
func OkVal(inner *Value) *Value  { return &Value{Kind: ValOk, Inner: inner} }
func ErrVal(inner *Value) *Value { return &Value{Kind: ValErr, Inner: inner} }
func PtrVal(addr int64) *Value   { return &Value{Kind: ValPtr, Int: addr} }

func ArrayVal(elems []*Value) *Value {
	return &Value{Kind: ValArray, Array: elems}
}

func MapVal(m *OrderedMap) *Value {
	return &Value{Kind: ValMap, Map: m}
}

func FnVal(fn *FnValue) *Value {
	return &Value{Kind: ValFn, Fn: fn}
}
