package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/joeabbey/morgoth/internal/lexer"
	"github.com/joeabbey/morgoth/internal/token"
)

// Precedence levels for Pratt parsing.
const (
	precLowest int = iota
	precAssign     // =
	precOr         // or
	precAnd        // and
	precEquality   // == === !=
	precComparison // < > <= >=
	precSum        // + -
	precProduct    // * / %
	precUnary      // - ! &
	precPostfix    // () [] . ? as
)

// Parser reads tokens from the lexer and produces an AST.
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
	buffered  []token.Token // tokens buffered by peekAhead, consumed before lexer
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns the list of parse errors.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d col %d: %s", p.curToken.Line, p.curToken.Col, msg))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	if len(p.buffered) > 0 {
		p.peekToken = p.buffered[0]
		p.buffered = p.buffered[1:]
	} else {
		p.peekToken = p.l.NextToken()
	}
}

// peekAhead returns the token n positions ahead of curToken (0 = curToken, 1 = peekToken, 2 = next, ...).
// It buffers tokens from the lexer as needed without advancing curToken or peekToken.
func (p *Parser) peekAhead(n int) token.Token {
	if n == 0 {
		return p.curToken
	}
	if n == 1 {
		return p.peekToken
	}
	// n >= 2: need buffered[n-2]
	idx := n - 2
	for len(p.buffered) <= idx {
		p.buffered = append(p.buffered, p.l.NextToken())
	}
	return p.buffered[idx]
}

func (p *Parser) curIs(t token.TokenType) bool  { return p.curToken.Type == t }
func (p *Parser) peekIs(t token.TokenType) bool { return p.peekToken.Type == t }

// expectPeek checks that peekToken is t, advances, returns true. Otherwise adds error.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekIs(t) {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected %s, got %s (%q)", t, p.peekToken.Type, p.peekToken.Literal))
	return false
}

// Parse parses the entire program and returns the AST.
func (p *Parser) Parse() *Program {
	prog := &Program{}
	for !p.curIs(token.EOF) {
		// Skip stray semicolons at top level.
		if p.curIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}
		item := p.parseItem()
		if item != nil {
			prog.Items = append(prog.Items, item)
		} else {
			p.nextToken()
		}
	}
	return prog
}

func (p *Parser) parseItem() Item {
	switch p.curToken.Type {
	case token.FN:
		return p.parseFnDecl()
	case token.EXTERN:
		return p.parseExternDecl()
	case token.LET:
		return p.parseLetStmt()
	case token.CONST:
		return p.parseConstStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.DECREE:
		return p.parseDecreeStmt()
	default:
		return p.parseExprStmt()
	}
}

// --- Declarations ---

func (p *Parser) parseFnDecl() *FnDecl {
	decl := &FnDecl{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	decl.Params = p.parseParamList()
	// curToken should be RPAREN
	if !p.curIs(token.RPAREN) {
		p.addError(fmt.Sprintf("expected ), got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past )
	body := p.parseBlockExpr()
	if body == nil {
		return nil
	}
	decl.Body = body
	return decl
}

func (p *Parser) parseExternDecl() *ExternDecl {
	decl := &ExternDecl{Token: p.curToken}
	if !p.expectPeek(token.FN) {
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	decl.Params = p.parseParamList()
	if !p.curIs(token.RPAREN) {
		p.addError(fmt.Sprintf("expected ), got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past )
	if p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return decl
}

// parseParamList parses parameter list. Called with curToken on (.
// Returns with curToken on ).
func (p *Parser) parseParamList() []Param {
	var params []Param
	p.nextToken() // move past (
	if p.curIs(token.RPAREN) {
		return params
	}
	for {
		if !p.curIs(token.IDENT) {
			p.addError(fmt.Sprintf("expected parameter name, got %s", p.curToken.Type))
			return params
		}
		param := Param{Name: p.curToken.Literal}
		if p.peekIs(token.COLON) {
			p.nextToken() // move to :
			p.nextToken() // move to type name
			param.Type = p.curToken.Literal
		}
		params = append(params, param)
		if !p.peekIs(token.COMMA) {
			break
		}
		p.nextToken() // move to comma
		p.nextToken() // move past comma to next param
	}
	p.nextToken() // advance to ) or next token
	return params
}

// --- Statements ---

func (p *Parser) parseStmt() Stmt {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStmt()
	case token.CONST:
		return p.parseConstStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.DECREE:
		return p.parseDecreeStmt()
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) parseLetStmt() *LetStmt {
	stmt := &LetStmt{Token: p.curToken}
	p.nextToken() // move past let
	// Allow keywords like "ok" and "err" as variable names.
	if !p.curIs(token.IDENT) && !p.curIs(token.OK) && !p.curIs(token.ERR) {
		p.addError(fmt.Sprintf("expected identifier after let, got %s (%q)", p.curToken.Type, p.curToken.Literal))
		return nil
	}
	stmt.Name = p.curToken.Literal
	if p.peekIs(token.COLON) {
		p.nextToken() // move to :
		p.nextToken() // move to type name
		stmt.TypeAnnotation = p.curToken.Literal
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken() // move past =
	stmt.Value = p.parseExpression(precLowest)
	if p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseConstStmt() *ConstStmt {
	stmt := &ConstStmt{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = p.curToken.Literal
	if p.peekIs(token.COLON) {
		p.nextToken() // move to :
		p.nextToken() // move to type name
		stmt.TypeAnnotation = p.curToken.Literal
	}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken() // move past =
	stmt.Value = p.parseExpression(precLowest)
	if p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStmt() *ReturnStmt {
	stmt := &ReturnStmt{Token: p.curToken}
	p.nextToken() // move past return
	stmt.Value = p.parseExpression(precLowest)
	if p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseDecreeStmt() *DecreeStmt {
	stmt := &DecreeStmt{Token: p.curToken}
	if !p.expectPeek(token.STRING) {
		return nil
	}
	stmt.Value = p.curToken.Literal
	p.nextToken() // move past string
	if p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExprStmt() *ExprStmt {
	stmt := &ExprStmt{Token: p.curToken}
	stmt.Expression = p.parseExpression(precLowest)
	if p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// --- Expressions (Pratt parser) ---
//
// Convention: after any prefix or infix parser returns, curToken is on the
// NEXT unconsumed token (the token immediately after the expression).
// So the Pratt loop checks curToken (not peekToken) for infix operators.

func (p *Parser) parseExpression(prec int) Expr {
	left := p.parsePrefixExpr()
	if left == nil {
		return nil
	}
	for {
		cp := p.curPrecedence()
		if cp <= prec {
			break
		}
		left = p.parseInfixExpr(left)
		if left == nil {
			return nil
		}
	}
	return left
}

func (p *Parser) curPrecedence() int {
	return tokenPrecedence(p.curToken.Type)
}

func tokenPrecedence(t token.TokenType) int {
	switch t {
	case token.ASSIGN:
		return precAssign
	case token.OR:
		return precOr
	case token.AND:
		return precAnd
	case token.EQ, token.STRICT_EQ, token.NEQ:
		return precEquality
	case token.LT, token.GT, token.LTE, token.GTE:
		return precComparison
	case token.PLUS, token.MINUS:
		return precSum
	case token.STAR, token.SLASH, token.PERCENT:
		return precProduct
	case token.LPAREN, token.LBRACKET, token.DOT, token.QUESTION, token.AS:
		return precPostfix
	default:
		return 0
	}
}

func (p *Parser) parsePrefixExpr() Expr {
	switch p.curToken.Type {
	case token.INT:
		return p.parseIntLit()
	case token.FLOAT:
		return p.parseFloatLit()
	case token.STRING:
		return p.parseStringLit()
	case token.TRUE, token.FALSE:
		return p.parseBoolLit()
	case token.NIL:
		return p.parseNilLit()
	case token.IDENT:
		return p.parseIdentExpr()
	case token.MINUS, token.BANG, token.AMP:
		return p.parseUnaryExpr()
	case token.LPAREN:
		return p.parseGroupedExpr()
	case token.LBRACKET:
		return p.parseArrayLitExpr()
	case token.LBRACE:
		return p.parseBlockOrMap()
	case token.IF:
		return p.parseIfExpr()
	case token.MATCH:
		return p.parseMatchExpr()
	case token.GUARD:
		return p.parseGuardExpr()
	case token.OK:
		return p.parseOkExpr()
	case token.ERR:
		return p.parseErrExpr()
	case token.SPEAK:
		return p.parseSpeakExpr()
	case token.SORRY:
		return p.parseSorryExpr()
	case token.DOOM:
		return p.parseDoomExpr()
	case token.CHANT:
		return p.parseChantExpr()
	case token.SPAWN:
		return p.parseSpawnExpr()
	case token.AWAIT_ALL:
		return p.parseAwaitAllExpr()
	default:
		p.addError(fmt.Sprintf("unexpected token %s (%q)", p.curToken.Type, p.curToken.Literal))
		return nil
	}
}

func (p *Parser) parseInfixExpr(left Expr) Expr {
	switch p.curToken.Type {
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.PERCENT,
		token.EQ, token.STRICT_EQ, token.NEQ,
		token.LT, token.GT, token.LTE, token.GTE,
		token.AND, token.OR:
		return p.parseBinaryExpr(left)
	case token.ASSIGN:
		return p.parseAssignExpr(left)
	case token.LPAREN:
		return p.parseCallExpr(left)
	case token.LBRACKET:
		return p.parseIndexExpr(left)
	case token.DOT:
		return p.parseDotExpr(left)
	case token.QUESTION:
		return p.parsePropagateExpr(left)
	case token.AS:
		return p.parseAsExpr(left)
	default:
		return left
	}
}

// --- Infix parsers ---
// curToken is on the operator when these are called.

func (p *Parser) parseBinaryExpr(left Expr) Expr {
	prec := p.curPrecedence()
	expr := &BinaryExpr{
		Token: p.curToken,
		Left:  left,
		Op:    p.curToken.Literal,
	}
	p.nextToken() // move past operator
	expr.Right = p.parseExpression(prec)
	return expr
}

func (p *Parser) parseAssignExpr(left Expr) Expr {
	tok := p.curToken // the = token
	p.nextToken()     // move past =
	value := p.parseExpression(precLowest)

	switch lhs := left.(type) {
	case *IdentExpr:
		return &AssignExpr{
			Token: tok,
			Name:  lhs.Name,
			Value: value,
		}
	case *IndexExpr:
		return &IndexAssignExpr{
			Token: tok,
			Left:  lhs.Left,
			Index: lhs.Index,
			Value: value,
		}
	case *DotExpr:
		return &DotAssignExpr{
			Token: tok,
			Left:  lhs.Left,
			Field: lhs.Field,
			Value: value,
		}
	default:
		p.addError("left side of assignment must be an identifier, index expression, or dot expression")
		return nil
	}
}

func (p *Parser) parseCallExpr(left Expr) Expr {
	expr := &CallExpr{
		Token:    p.curToken,
		Function: left,
	}
	expr.Args = p.parseExprList(token.RPAREN)
	return expr
}

func (p *Parser) parseIndexExpr(left Expr) Expr {
	expr := &IndexExpr{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken() // move past [
	expr.Index = p.parseExpression(precLowest)
	if !p.curIs(token.RBRACKET) {
		p.addError(fmt.Sprintf("expected ], got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past ]
	return expr
}

func (p *Parser) parseDotExpr(left Expr) Expr {
	expr := &DotExpr{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken() // move past .
	if !p.curIs(token.IDENT) {
		p.addError(fmt.Sprintf("expected identifier after ., got %s", p.curToken.Type))
		return nil
	}
	expr.Field = p.curToken.Literal
	p.nextToken() // move past field name
	return expr
}

func (p *Parser) parsePropagateExpr(left Expr) Expr {
	expr := &PropagateExpr{
		Token: p.curToken,
		Inner: left,
	}
	p.nextToken() // move past ?
	return expr
}

func (p *Parser) parseAsExpr(left Expr) Expr {
	tok := p.curToken
	p.nextToken() // move past 'as'
	typeName := p.curToken.Literal
	p.nextToken() // move past type name
	return &AsExpr{
		Token:    tok,
		Left:     left,
		TypeName: typeName,
	}
}

// --- Prefix parsers ---
// All leave curToken on the next unconsumed token after the expression.

func (p *Parser) parseIntLit() Expr {
	lit := p.curToken.Literal
	cleaned := strings.ReplaceAll(lit, "_", "")
	var val int64
	var err error
	if strings.HasPrefix(cleaned, "0x") || strings.HasPrefix(cleaned, "0X") {
		val, err = strconv.ParseInt(cleaned[2:], 16, 64)
	} else {
		val, err = strconv.ParseInt(cleaned, 10, 64)
	}
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as integer: %s", lit, err))
		return nil
	}
	expr := &IntLitExpr{Token: p.curToken, Value: val}
	p.nextToken()
	return expr
}

func (p *Parser) parseFloatLit() Expr {
	lit := p.curToken.Literal
	cleaned := strings.ReplaceAll(lit, "_", "")
	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as float: %s", lit, err))
		return nil
	}
	expr := &FloatLitExpr{Token: p.curToken, Value: val}
	p.nextToken()
	return expr
}

func (p *Parser) parseStringLit() Expr {
	expr := &StringLitExpr{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()
	return expr
}

func (p *Parser) parseBoolLit() Expr {
	expr := &BoolLitExpr{Token: p.curToken, Value: p.curIs(token.TRUE)}
	p.nextToken()
	return expr
}

func (p *Parser) parseNilLit() Expr {
	expr := &NilLitExpr{Token: p.curToken}
	p.nextToken()
	return expr
}

func (p *Parser) parseIdentExpr() Expr {
	expr := &IdentExpr{Token: p.curToken, Name: p.curToken.Literal}
	p.nextToken()
	return expr
}

func (p *Parser) parseUnaryExpr() Expr {
	expr := &UnaryExpr{Token: p.curToken, Op: p.curToken.Literal}
	p.nextToken() // move past operator
	expr.Right = p.parseExpression(precUnary)
	return expr
}

func (p *Parser) parseGroupedExpr() Expr {
	p.nextToken() // skip (
	expr := p.parseExpression(precLowest)
	if !p.curIs(token.RPAREN) {
		p.addError(fmt.Sprintf("expected ), got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // skip )
	return expr
}

func (p *Parser) parseArrayLitExpr() Expr {
	expr := &ArrayLitExpr{Token: p.curToken}
	expr.Elements = p.parseExprList(token.RBRACKET)
	return expr
}

// parseExprList parses a comma-separated list of expressions.
// curToken is on the opening delimiter (e.g., ( or [).
// Returns with curToken on the token AFTER the closing delimiter.
func (p *Parser) parseExprList(end token.TokenType) []Expr {
	var list []Expr
	p.nextToken() // move past opening delimiter
	if p.curIs(end) {
		p.nextToken() // move past closing delimiter
		return list
	}
	list = append(list, p.parseExpression(precLowest))
	for p.curIs(token.COMMA) {
		p.nextToken() // skip comma
		if p.curIs(end) {
			break // trailing comma
		}
		list = append(list, p.parseExpression(precLowest))
	}
	if !p.curIs(end) {
		p.addError(fmt.Sprintf("expected %s, got %s", end, p.curToken.Type))
		return list
	}
	p.nextToken() // move past closing delimiter
	return list
}

// --- Block / Map ---

func (p *Parser) parseBlockOrMap() Expr {
	if p.isMapLiteral() {
		return p.parseMapLitExpr()
	}
	return p.parseBlockExpr()
}

// isMapLiteral peeks ahead to decide if { starts a map literal.
// Map: { STRING : ... } or { IDENT/OK/ERR : ... } or { INT/FLOAT/BOOL/NIL : ... }
func (p *Parser) isMapLiteral() bool {
	if p.peekIs(token.STRING) {
		return true
	}
	// For all other key types, check if two tokens ahead is COLON.
	switch p.peekToken.Type {
	case token.IDENT, token.OK, token.ERR,
		token.INT, token.FLOAT, token.TRUE, token.FALSE, token.NIL:
		return p.peekAhead(2).Type == token.COLON
	}
	return false
}

func (p *Parser) parseMapLitExpr() Expr {
	expr := &MapLitExpr{Token: p.curToken}
	p.nextToken() // move past {
	for !p.curIs(token.RBRACE) && !p.curIs(token.EOF) {
		key := p.parseExpression(precLowest)
		if !p.curIs(token.COLON) {
			p.addError(fmt.Sprintf("expected :, got %s", p.curToken.Type))
			return expr
		}
		p.nextToken() // move past :
		value := p.parseExpression(precLowest)
		expr.Pairs = append(expr.Pairs, MapPair{Key: key, Value: value})
		if p.curIs(token.COMMA) {
			p.nextToken()
		} else if p.curIs(token.SEMICOLON) {
			p.nextToken()
		}
	}
	if !p.curIs(token.RBRACE) {
		p.addError(fmt.Sprintf("expected }, got %s", p.curToken.Type))
		return expr
	}
	p.nextToken() // move past }
	return expr
}

// parseBlockExpr parses { stmts... [finalExpr] }.
// curToken must be on {. Returns with curToken past }.
func (p *Parser) parseBlockExpr() *BlockExpr {
	if !p.curIs(token.LBRACE) {
		p.addError(fmt.Sprintf("expected {, got %s (%q)", p.curToken.Type, p.curToken.Literal))
		return nil
	}
	block := &BlockExpr{Token: p.curToken}
	p.nextToken() // move past {

	for !p.curIs(token.RBRACE) && !p.curIs(token.EOF) {
		if p.curIs(token.LET) || p.curIs(token.CONST) || p.curIs(token.RETURN) || p.curIs(token.DECREE) {
			stmt := p.parseStmt()
			if stmt != nil {
				block.Stmts = append(block.Stmts, stmt)
			}
			continue
		}

		expr := p.parseExpression(precLowest)
		if expr == nil {
			p.nextToken()
			continue
		}

		if p.curIs(token.SEMICOLON) {
			block.Stmts = append(block.Stmts, &ExprStmt{Expression: expr})
			p.nextToken() // consume ;
		} else if p.curIs(token.RBRACE) {
			block.FinalExpr = expr
		} else {
			block.Stmts = append(block.Stmts, &ExprStmt{Expression: expr})
		}
	}

	if !p.curIs(token.RBRACE) {
		p.addError("expected }")
		return block
	}
	p.nextToken() // move past }
	return block
}

// --- Keyword expression parsers ---

func (p *Parser) parseIfExpr() Expr {
	expr := &IfExpr{Token: p.curToken}
	p.nextToken() // move past if
	expr.Condition = p.parseExpression(precLowest)

	then := p.parseBlockExpr()
	if then == nil {
		return nil
	}
	expr.Then = then

	if p.curIs(token.ELSE) {
		p.nextToken() // move past else
		if p.curIs(token.IF) {
			expr.Else = p.parseIfExpr()
		} else if p.curIs(token.LBRACE) {
			expr.Else = p.parseBlockExpr()
		} else {
			// Bare expression after else — wrap in an implicit block.
			elseExpr := p.parseExpression(precLowest)
			expr.Else = &BlockExpr{FinalExpr: elseExpr}
		}
	}
	return expr
}

func (p *Parser) parseMatchExpr() Expr {
	expr := &MatchExpr{Token: p.curToken}
	p.nextToken() // move past match
	expr.Subject = p.parseExpression(precLowest)

	if !p.curIs(token.LBRACE) {
		p.addError(fmt.Sprintf("expected { after match subject, got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past {

	for !p.curIs(token.RBRACE) && !p.curIs(token.EOF) {
		arm := p.parseMatchArm()
		expr.Arms = append(expr.Arms, arm)
	}
	if p.curIs(token.RBRACE) {
		p.nextToken() // move past }
	}
	return expr
}

func (p *Parser) parseMatchArm() MatchArm {
	arm := MatchArm{}
	arm.Pattern = p.parsePattern()

	if !p.curIs(token.ARROW) {
		p.addError(fmt.Sprintf("expected =>, got %s (%q)", p.curToken.Type, p.curToken.Literal))
		return arm
	}
	p.nextToken() // move past =>

	arm.Body = p.parseExpression(precLowest)

	if p.curIs(token.COMMA) || p.curIs(token.SEMICOLON) {
		p.nextToken()
	}
	return arm
}

func (p *Parser) parsePattern() Pattern {
	// _ is wildcard
	if p.curIs(token.IDENT) && p.curToken.Literal == "_" {
		pat := &WildcardPattern{Token: p.curToken}
		p.nextToken()
		return p.maybeGuardedPattern(pat)
	}

	// ok(v) / err(e) destructuring patterns in match arms
	if (p.curIs(token.OK) || p.curIs(token.ERR)) && p.peekIs(token.LPAREN) {
		name := p.curToken.Literal
		tok := p.curToken
		p.nextToken() // skip ok/err
		p.nextToken() // skip (
		inner := ""
		if p.curIs(token.IDENT) {
			inner = p.curToken.Literal
			p.nextToken()
		}
		if p.curIs(token.RPAREN) {
			p.nextToken() // skip )
		}
		pat := &IdentPattern{Token: tok, Name: name + "(" + inner + ")"}
		return p.maybeGuardedPattern(pat)
	}

	// Literal patterns: int, float, string, bool, nil
	if p.curIs(token.INT) || p.curIs(token.FLOAT) || p.curIs(token.STRING) ||
		p.curIs(token.TRUE) || p.curIs(token.FALSE) || p.curIs(token.NIL) {
		expr := p.parsePrefixExpr()
		pat := &LiteralPattern{Token: p.curToken, Value: expr}
		return p.maybeGuardedPattern(pat)
	}

	// Negative literal: -int or -float
	if p.curIs(token.MINUS) && (p.peekIs(token.INT) || p.peekIs(token.FLOAT)) {
		expr := p.parseUnaryExpr()
		pat := &LiteralPattern{Token: p.curToken, Value: expr}
		return p.maybeGuardedPattern(pat)
	}

	// Ident or typed pattern (ident : type)
	if p.curIs(token.IDENT) {
		tok := p.curToken
		name := p.curToken.Literal
		p.nextToken()
		if p.curIs(token.COLON) {
			p.nextToken() // skip :
			typeName := p.curToken.Literal
			p.nextToken() // skip type name
			pat := &TypedPattern{Token: tok, Name: name, TypeName: typeName}
			return p.maybeGuardedPattern(pat)
		}
		pat := &IdentPattern{Token: tok, Name: name}
		return p.maybeGuardedPattern(pat)
	}

	p.addError(fmt.Sprintf("unexpected token in pattern: %s (%q)", p.curToken.Type, p.curToken.Literal))
	p.nextToken()
	return &WildcardPattern{Token: p.curToken}
}

func (p *Parser) maybeGuardedPattern(inner Pattern) Pattern {
	if p.curIs(token.IF) {
		tok := p.curToken
		p.nextToken() // move past if
		guard := p.parseExpression(precLowest)
		return &GuardedPattern{Token: tok, Inner: inner, Guard: guard}
	}
	return inner
}

func (p *Parser) parseGuardExpr() Expr {
	expr := &GuardExpr{Token: p.curToken}
	p.nextToken() // move past guard
	expr.Condition = p.parseExpression(precLowest)
	if !p.curIs(token.ELSE) {
		p.addError(fmt.Sprintf("expected else after guard condition, got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past else
	expr.ElseBody = p.parseExpression(precLowest)
	return expr
}

func (p *Parser) parseOkExpr() Expr {
	tok := p.curToken
	if !p.peekIs(token.LPAREN) {
		expr := &IdentExpr{Token: tok, Name: tok.Literal}
		p.nextToken()
		return expr
	}
	p.nextToken() // move to (
	p.nextToken() // move past (
	inner := p.parseExpression(precLowest)
	if !p.curIs(token.RPAREN) {
		p.addError(fmt.Sprintf("expected ) in ok(), got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past )
	return &OkExpr{Token: tok, Inner: inner}
}

func (p *Parser) parseErrExpr() Expr {
	tok := p.curToken
	if !p.peekIs(token.LPAREN) {
		expr := &IdentExpr{Token: tok, Name: tok.Literal}
		p.nextToken()
		return expr
	}
	p.nextToken() // move to (
	p.nextToken() // move past (
	inner := p.parseExpression(precLowest)
	if !p.curIs(token.RPAREN) {
		p.addError(fmt.Sprintf("expected ) in err(), got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past )
	return &ErrExpr{Token: tok, Inner: inner}
}

func (p *Parser) parseSpeakExpr() Expr {
	tok := p.curToken
	p.nextToken() // move past speak
	value := p.parseExpression(precLowest)
	var elseBody Expr
	if p.curIs(token.ELSE) {
		p.nextToken() // move past else
		elseBody = p.parseExpression(precLowest)
	}
	return &SpeakExpr{Token: tok, Value: value, ElseBody: elseBody}
}

func (p *Parser) parseSorryExpr() Expr {
	tok := p.curToken
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // move past (
	if !p.curIs(token.IDENT) {
		p.addError(fmt.Sprintf("expected identifier in sorry(), got %s", p.curToken.Type))
		return nil
	}
	name := p.curToken.Literal
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	p.nextToken() // move past )
	return &SorryExpr{Token: tok, Name: name}
}

func (p *Parser) parseDoomExpr() Expr {
	tok := p.curToken
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // move past (
	msg := p.parseExpression(precLowest)
	if !p.curIs(token.RPAREN) {
		p.addError(fmt.Sprintf("expected ) in doom(), got %s", p.curToken.Type))
		return nil
	}
	p.nextToken() // move past )
	return &DoomExpr{Token: tok, Message: msg}
}

func (p *Parser) parseChantExpr() Expr {
	tok := p.curToken
	p.nextToken() // move past chant
	name := p.parseExpression(precLowest)
	return &ChantExpr{Token: tok, Name: name}
}

func (p *Parser) parseSpawnExpr() Expr {
	tok := p.curToken
	p.nextToken() // move past spawn
	body := p.parseBlockExpr()
	if body == nil {
		return nil
	}
	return &SpawnExpr{Token: tok, Body: body}
}

func (p *Parser) parseAwaitAllExpr() Expr {
	tok := p.curToken
	if p.peekIs(token.LPAREN) {
		p.nextToken() // move to (
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		p.nextToken() // move past )
	} else {
		p.nextToken() // move past await_all keyword
	}
	return &AwaitAllExpr{Token: tok}
}
