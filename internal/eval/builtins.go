package eval

import (
	"os"
	"unicode/utf8"
)

// callBuiltin dispatches to built-in functions that are invoked via CallExpr
// (as opposed to speak/doom/sorry/chant which are special AST nodes).
// Returns (result, true) if name is a builtin, or (nil, false) otherwise.
func (ev *Evaluator) callBuiltin(name string, args []*Value) (*Value, bool, error) {
	switch name {
	case "len":
		return ev.builtinLen(args)
	case "malloc":
		return PtrVal(0), true, nil
	case "free":
		return OkVal(NilVal()), true, nil
	case "read":
		return StrVal(""), true, nil
	case "write":
		return OkVal(NilVal()), true, nil
	case "read_file":
		return ev.builtinReadFile(args)
	case "parse_toml":
		return ErrVal(StrVal("not implemented")), true, nil
	default:
		return nil, false, nil
	}
}

func (ev *Evaluator) builtinLen(args []*Value) (*Value, bool, error) {
	if len(args) != 1 {
		return nil, true, &DoomError{Message: "len() takes exactly 1 argument"}
	}
	switch args[0].Kind {
	case ValArray:
		return IntVal(int64(len(args[0].Array))), true, nil
	case ValStr:
		return IntVal(int64(utf8.RuneCountInString(args[0].Str))), true, nil
	case ValMap:
		return IntVal(int64(args[0].Map.Len())), true, nil
	default:
		return nil, true, &DoomError{Message: "len() argument must be array, string, or map"}
	}
}

func (ev *Evaluator) builtinReadFile(args []*Value) (*Value, bool, error) {
	if len(args) != 1 || args[0].Kind != ValStr {
		return ErrVal(StrVal("read_file() takes exactly 1 string argument")), true, nil
	}
	data, err := os.ReadFile(args[0].Str)
	if err != nil {
		return ErrVal(StrVal(err.Error())), true, nil
	}
	return OkVal(StrVal(string(data))), true, nil
}
