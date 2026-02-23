package parser

import "github.com/joeabbey/morgoth/internal/token"

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
}

// Stmt is a statement node.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is an expression node.
type Expr interface {
	Node
	exprNode()
}

// Pattern is a match arm pattern.
type Pattern interface {
	Node
	patternNode()
}

// Item is something that can appear at the top level: fn decl, extern decl, or stmt.
type Item interface {
	Node
	itemNode()
}

// Program is the root AST node.
type Program struct {
	Items []Item
}

func (p *Program) TokenLiteral() string {
	if len(p.Items) > 0 {
		return p.Items[0].TokenLiteral()
	}
	return ""
}

// --- Declarations ---

// FnDecl represents a function declaration: fn name(params) { body }
type FnDecl struct {
	Token  token.Token // the FN token
	Name   string
	Params []Param
	Body   *BlockExpr
}

func (d *FnDecl) TokenLiteral() string { return d.Token.Literal }
func (d *FnDecl) itemNode()            {}

// Param is a function parameter.
type Param struct {
	Name string
	Type string // optional type annotation
}

// ExternDecl represents: extern fn name(params);
type ExternDecl struct {
	Token  token.Token // the EXTERN token
	Name   string
	Params []Param
}

func (d *ExternDecl) TokenLiteral() string { return d.Token.Literal }
func (d *ExternDecl) itemNode()            {}

// --- Statements ---

// LetStmt represents: let name [: type] = value;
type LetStmt struct {
	Token          token.Token
	Name           string
	TypeAnnotation string
	Value          Expr
}

func (s *LetStmt) TokenLiteral() string { return s.Token.Literal }
func (s *LetStmt) stmtNode()            {}
func (s *LetStmt) itemNode()            {}

// ConstStmt represents: const name [: type] = value;
type ConstStmt struct {
	Token          token.Token
	Name           string
	TypeAnnotation string
	Value          Expr
}

func (s *ConstStmt) TokenLiteral() string { return s.Token.Literal }
func (s *ConstStmt) stmtNode()            {}
func (s *ConstStmt) itemNode()            {}

// ReturnStmt represents: return expr;
type ReturnStmt struct {
	Token token.Token
	Value Expr
}

func (s *ReturnStmt) TokenLiteral() string { return s.Token.Literal }
func (s *ReturnStmt) stmtNode()            {}
func (s *ReturnStmt) itemNode()            {}

// DecreeStmt represents: decree "string";
type DecreeStmt struct {
	Token token.Token
	Value string
}

func (s *DecreeStmt) TokenLiteral() string { return s.Token.Literal }
func (s *DecreeStmt) stmtNode()            {}
func (s *DecreeStmt) itemNode()            {}

// ExprStmt wraps an expression used as a statement.
type ExprStmt struct {
	Token      token.Token
	Expression Expr
}

func (s *ExprStmt) TokenLiteral() string { return s.Token.Literal }
func (s *ExprStmt) stmtNode()            {}
func (s *ExprStmt) itemNode()            {}

// --- Expressions ---

// IntLitExpr represents an integer literal.
type IntLitExpr struct {
	Token token.Token
	Value int64
}

func (e *IntLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *IntLitExpr) exprNode()            {}

// FloatLitExpr represents a floating-point literal.
type FloatLitExpr struct {
	Token token.Token
	Value float64
}

func (e *FloatLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *FloatLitExpr) exprNode()            {}

// StringLitExpr represents a string literal.
type StringLitExpr struct {
	Token token.Token
	Value string
}

func (e *StringLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *StringLitExpr) exprNode()            {}

// BoolLitExpr represents true or false.
type BoolLitExpr struct {
	Token token.Token
	Value bool
}

func (e *BoolLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *BoolLitExpr) exprNode()            {}

// NilLitExpr represents nil.
type NilLitExpr struct {
	Token token.Token
}

func (e *NilLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *NilLitExpr) exprNode()            {}

// IdentExpr represents an identifier reference.
type IdentExpr struct {
	Token token.Token
	Name  string
}

func (e *IdentExpr) TokenLiteral() string { return e.Token.Literal }
func (e *IdentExpr) exprNode()            {}

// ArrayLitExpr represents [elem, elem, ...].
type ArrayLitExpr struct {
	Token    token.Token // the LBRACKET
	Elements []Expr
}

func (e *ArrayLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *ArrayLitExpr) exprNode()            {}

// MapPair is a key-value pair in a map literal.
type MapPair struct {
	Key   Expr
	Value Expr
}

// MapLitExpr represents { key: value, ... }.
type MapLitExpr struct {
	Token token.Token // the LBRACE
	Pairs []MapPair
}

func (e *MapLitExpr) TokenLiteral() string { return e.Token.Literal }
func (e *MapLitExpr) exprNode()            {}

// BinaryExpr represents left op right.
type BinaryExpr struct {
	Token token.Token
	Left  Expr
	Op    string
	Right Expr
}

func (e *BinaryExpr) TokenLiteral() string { return e.Token.Literal }
func (e *BinaryExpr) exprNode()            {}

// UnaryExpr represents op right (prefix).
type UnaryExpr struct {
	Token token.Token
	Op    string
	Right Expr
}

func (e *UnaryExpr) TokenLiteral() string { return e.Token.Literal }
func (e *UnaryExpr) exprNode()            {}

// AssignExpr represents name = value.
type AssignExpr struct {
	Token token.Token
	Name  string
	Value Expr
}

func (e *AssignExpr) TokenLiteral() string { return e.Token.Literal }
func (e *AssignExpr) exprNode()            {}

// CallExpr represents function(args...).
type CallExpr struct {
	Token    token.Token // the LPAREN
	Function Expr
	Args     []Expr
}

func (e *CallExpr) TokenLiteral() string { return e.Token.Literal }
func (e *CallExpr) exprNode()            {}

// IndexExpr represents left[index].
type IndexExpr struct {
	Token token.Token // the LBRACKET
	Left  Expr
	Index Expr
}

func (e *IndexExpr) TokenLiteral() string { return e.Token.Literal }
func (e *IndexExpr) exprNode()            {}

// DotExpr represents left.field.
type DotExpr struct {
	Token token.Token // the DOT
	Left  Expr
	Field string
}

func (e *DotExpr) TokenLiteral() string { return e.Token.Literal }
func (e *DotExpr) exprNode()            {}

// PropagateExpr represents expr? (error propagation).
type PropagateExpr struct {
	Token token.Token // the QUESTION
	Inner Expr
}

func (e *PropagateExpr) TokenLiteral() string { return e.Token.Literal }
func (e *PropagateExpr) exprNode()            {}

// IfExpr represents: if cond { ... } else { ... }
type IfExpr struct {
	Token     token.Token // the IF token
	Condition Expr
	Then      *BlockExpr
	Else      Expr // *BlockExpr or *IfExpr, or nil
}

func (e *IfExpr) TokenLiteral() string { return e.Token.Literal }
func (e *IfExpr) exprNode()            {}

// MatchArm is a single arm in a match expression.
type MatchArm struct {
	Pattern Pattern
	Body    Expr
}

// MatchExpr represents: match subject { arms... }
type MatchExpr struct {
	Token   token.Token // the MATCH token
	Subject Expr
	Arms    []MatchArm
}

func (e *MatchExpr) TokenLiteral() string { return e.Token.Literal }
func (e *MatchExpr) exprNode()            {}

// GuardExpr represents: guard condition else body
type GuardExpr struct {
	Token     token.Token // the GUARD token
	Condition Expr
	ElseBody  Expr
}

func (e *GuardExpr) TokenLiteral() string { return e.Token.Literal }
func (e *GuardExpr) exprNode()            {}

// BlockExpr represents { stmts... [final_expr] }
type BlockExpr struct {
	Token     token.Token // the LBRACE
	Stmts     []Stmt
	FinalExpr Expr // optional trailing expression (implicit return)
}

func (e *BlockExpr) TokenLiteral() string { return e.Token.Literal }
func (e *BlockExpr) exprNode()            {}

// OkExpr represents ok(expr).
type OkExpr struct {
	Token token.Token
	Inner Expr
}

func (e *OkExpr) TokenLiteral() string { return e.Token.Literal }
func (e *OkExpr) exprNode()            {}

// ErrExpr represents err(expr).
type ErrExpr struct {
	Token token.Token
	Inner Expr
}

func (e *ErrExpr) TokenLiteral() string { return e.Token.Literal }
func (e *ErrExpr) exprNode()            {}

// AsExpr represents expr as type (type coercion).
type AsExpr struct {
	Token    token.Token // the AS token
	Left     Expr
	TypeName string
}

func (e *AsExpr) TokenLiteral() string { return e.Token.Literal }
func (e *AsExpr) exprNode()            {}

// SpeakExpr represents: speak expr [else expr]
type SpeakExpr struct {
	Token    token.Token // the SPEAK token
	Value    Expr
	ElseBody Expr // optional
}

func (e *SpeakExpr) TokenLiteral() string { return e.Token.Literal }
func (e *SpeakExpr) exprNode()            {}

// SorryExpr represents: sorry(ident)
type SorryExpr struct {
	Token token.Token
	Name  string
}

func (e *SorryExpr) TokenLiteral() string { return e.Token.Literal }
func (e *SorryExpr) exprNode()            {}

// DoomExpr represents: doom(expr)
type DoomExpr struct {
	Token   token.Token
	Message Expr
}

func (e *DoomExpr) TokenLiteral() string { return e.Token.Literal }
func (e *DoomExpr) exprNode()            {}

// ChantExpr represents: chant expr
type ChantExpr struct {
	Token token.Token
	Name  Expr
}

func (e *ChantExpr) TokenLiteral() string { return e.Token.Literal }
func (e *ChantExpr) exprNode()            {}

// --- Patterns ---

// WildcardPattern matches anything: _
type WildcardPattern struct {
	Token token.Token
}

func (p *WildcardPattern) TokenLiteral() string { return p.Token.Literal }
func (p *WildcardPattern) patternNode()          {}

// LiteralPattern matches a literal value.
type LiteralPattern struct {
	Token token.Token
	Value Expr
}

func (p *LiteralPattern) TokenLiteral() string { return p.Token.Literal }
func (p *LiteralPattern) patternNode()          {}

// IdentPattern matches and binds a name.
type IdentPattern struct {
	Token token.Token
	Name  string
}

func (p *IdentPattern) TokenLiteral() string { return p.Token.Literal }
func (p *IdentPattern) patternNode()          {}

// TypedPattern matches with a type annotation: name: type
type TypedPattern struct {
	Token    token.Token
	Name     string
	TypeName string
}

func (p *TypedPattern) TokenLiteral() string { return p.Token.Literal }
func (p *TypedPattern) patternNode()          {}

// GuardedPattern adds a guard condition to a pattern: pattern if expr
type GuardedPattern struct {
	Token token.Token
	Inner Pattern
	Guard Expr
}

func (p *GuardedPattern) TokenLiteral() string { return p.Token.Literal }
func (p *GuardedPattern) patternNode()          {}
