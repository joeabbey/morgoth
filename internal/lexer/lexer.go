package lexer

import (
	"strings"

	"github.com/joeabbey/morgoth/internal/token"
)

// Lexer scans Morgoth source code into tokens.
type Lexer struct {
	input   string
	pos     int  // current position in input (points to current char)
	readPos int  // current reading position (after current char)
	ch      byte // current char under examination
	line    int
	col     int

	// For semicolon insertion: track the last non-whitespace token emitted.
	lastToken token.Token
	// pendingSemicolon is set when we detect a newline boundary requiring
	// semicolon insertion; it will be emitted before the next real token.
	pendingSemicolon *token.Token
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
		// Initialize lastToken to EOF so that initial newlines don't
		// trigger spurious semicolon insertion (since INT is iota 0).
		lastToken: token.Token{Type: token.EOF},
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) peekCharAt(offset int) byte {
	idx := l.readPos + offset
	if idx >= len(l.input) {
		return 0
	}
	return l.input[idx]
}

// skipWhitespaceAndComments skips whitespace (spaces, tabs, \r) and comments.
// It does NOT skip newlines — those are significant for semicolon insertion.
// Returns true if a newline was crossed.
func (l *Lexer) skipWhitespaceAndComments() bool {
	sawNewline := false
	for {
		switch {
		case l.ch == ' ' || l.ch == '\t' || l.ch == '\r':
			l.readChar()
		case l.ch == '\n':
			sawNewline = true
			l.line++
			l.col = 0
			l.readChar()
		case l.ch == '#':
			if l.peekChar() == '{' {
				l.skipBlockComment()
			} else {
				l.skipLineComment()
			}
		default:
			return sawNewline
		}
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipBlockComment() {
	// consume the '#{'
	l.readChar() // skip '#'
	l.readChar() // skip '{'
	depth := 1
	for depth > 0 && l.ch != 0 {
		if l.ch == '#' && l.peekChar() == '{' {
			if depth < 2 {
				depth++
				l.readChar()
				l.readChar()
			} else {
				// At depth >= 2, treat #{ as regular comment text (cap at 2)
				l.readChar()
			}
		} else if l.ch == '}' && l.peekChar() == '#' {
			depth--
			l.readChar()
			l.readChar()
		} else {
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			l.readChar()
		}
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	// If we have a pending semicolon from newline insertion, emit it first.
	if l.pendingSemicolon != nil {
		tok := *l.pendingSemicolon
		l.pendingSemicolon = nil
		l.lastToken = tok
		return tok
	}

	sawNewline := l.skipWhitespaceAndComments()

	// Check for semicolon insertion:
	// If we crossed a newline, the last token triggers semicolon insertion,
	// and the upcoming token starts a statement (or is EOF).
	if sawNewline && token.SemicolonTrigger(l.lastToken.Type) {
		// Peek at what comes next to see if it starts a statement or is EOF.
		if l.ch == 0 || l.nextTokenStartsStatement() {
			semi := token.Token{
				Type:    token.SEMICOLON,
				Literal: ";",
				Line:    l.line,
				Col:     l.col,
			}
			l.pendingSemicolon = &semi
			// Recurse to emit the semicolon.
			return l.NextToken()
		}
	}

	var tok token.Token
	tok.Line = l.line
	tok.Col = l.col

	switch {
	case l.ch == 0:
		// Check for trailing semicolon insertion at EOF.
		if token.SemicolonTrigger(l.lastToken.Type) {
			tok.Type = token.SEMICOLON
			tok.Literal = ";"
			l.lastToken = tok
			// Set lastToken to SEMICOLON so next call yields EOF.
			return tok
		}
		tok.Type = token.EOF
		tok.Literal = ""
		l.lastToken = tok
		return tok

	case l.ch == '+':
		tok = l.makeToken(token.PLUS, "+")
		l.readChar()

	case l.ch == '-':
		tok = l.makeToken(token.MINUS, "-")
		l.readChar()

	case l.ch == '*':
		tok = l.makeToken(token.STAR, "*")
		l.readChar()

	case l.ch == '/':
		tok = l.makeToken(token.SLASH, "/")
		l.readChar()

	case l.ch == '%':
		tok = l.makeToken(token.PERCENT, "%")
		l.readChar()

	case l.ch == '&':
		tok = l.makeToken(token.AMP, "&")
		l.readChar()

	case l.ch == '(':
		tok = l.makeToken(token.LPAREN, "(")
		l.readChar()

	case l.ch == ')':
		tok = l.makeToken(token.RPAREN, ")")
		l.readChar()

	case l.ch == '[':
		tok = l.makeToken(token.LBRACKET, "[")
		l.readChar()

	case l.ch == ']':
		tok = l.makeToken(token.RBRACKET, "]")
		l.readChar()

	case l.ch == '{':
		tok = l.makeToken(token.LBRACE, "{")
		l.readChar()

	case l.ch == '}':
		tok = l.makeToken(token.RBRACE, "}")
		l.readChar()

	case l.ch == ',':
		tok = l.makeToken(token.COMMA, ",")
		l.readChar()

	case l.ch == ';':
		tok = l.makeToken(token.SEMICOLON, ";")
		l.readChar()

	case l.ch == ':':
		tok = l.makeToken(token.COLON, ":")
		l.readChar()

	case l.ch == '.':
		tok = l.makeToken(token.DOT, ".")
		l.readChar()

	case l.ch == '?':
		tok = l.makeToken(token.QUESTION, "?")
		l.readChar()

	case l.ch == '=':
		if l.peekChar() == '=' && l.peekCharAt(1) == '=' {
			// ===
			tok = l.makeToken(token.STRICT_EQ, "===")
			l.readChar() // skip first =
			l.readChar() // skip second =
			l.readChar() // skip third =
		} else if l.peekChar() == '=' {
			// ==
			tok = l.makeToken(token.EQ, "==")
			l.readChar()
			l.readChar()
		} else if l.peekChar() == '>' {
			tok = l.makeToken(token.ARROW, "=>")
			l.readChar()
			l.readChar()
		} else {
			tok = l.makeToken(token.ASSIGN, "=")
			l.readChar()
		}

	case l.ch == '!':
		if l.peekChar() == '=' {
			tok = l.makeToken(token.NEQ, "!=")
			l.readChar()
			l.readChar()
		} else {
			tok = l.makeToken(token.BANG, "!")
			l.readChar()
		}

	case l.ch == '<':
		if l.peekChar() == '=' {
			tok = l.makeToken(token.LTE, "<=")
			l.readChar()
			l.readChar()
		} else {
			tok = l.makeToken(token.LT, "<")
			l.readChar()
		}

	case l.ch == '>':
		if l.peekChar() == '=' {
			tok = l.makeToken(token.GTE, ">=")
			l.readChar()
			l.readChar()
		} else {
			tok = l.makeToken(token.GT, ">")
			l.readChar()
		}

	case l.ch == '"':
		var ok bool
		tok.Literal, ok = l.readString()
		if ok {
			tok.Type = token.STRING
		} else {
			tok.Type = token.ILLEGAL
		}

	case isDigit(l.ch):
		tok.Type, tok.Literal = l.readNumber()

	case isLetter(l.ch):
		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdent(tok.Literal)

	default:
		tok = l.makeToken(token.ILLEGAL, string(l.ch))
		l.readChar()
	}

	l.lastToken = tok
	return tok
}

// Tokenize returns all tokens from the input until EOF (inclusive).
func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) makeToken(tt token.TokenType, literal string) token.Token {
	return token.Token{
		Type:    tt,
		Literal: literal,
		Line:    l.line,
		Col:     l.col,
	}
}

func (l *Lexer) readString() (string, bool) {
	var sb strings.Builder
	l.readChar() // skip opening quote
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '0':
				sb.WriteByte(0)
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			default:
				// Unknown escape: include as-is.
				sb.WriteByte('\\')
				sb.WriteByte(l.ch)
			}
		} else {
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			sb.WriteByte(l.ch)
		}
		l.readChar()
	}
	if l.ch == '"' {
		l.readChar() // skip closing quote
		return sb.String(), true
	}
	// Unterminated string
	return sb.String(), false
}

func (l *Lexer) readNumber() (token.TokenType, string) {
	start := l.pos
	isFloat := false

	// Check for hex
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar() // '0'
		l.readChar() // 'x'
		for isHexDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
		return token.INT, l.input[start:l.pos]
	}

	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Check for float
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // skip '.'
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	if isFloat {
		return token.FLOAT, l.input[start:l.pos]
	}
	return token.INT, l.input[start:l.pos]
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

// nextTokenStartsStatement peeks ahead to see if the next non-whitespace
// character(s) form a keyword that starts a statement.
func (l *Lexer) nextTokenStartsStatement() bool {
	if l.ch == 0 {
		return true // EOF triggers insertion
	}
	if !isLetter(l.ch) {
		return false
	}
	// Peek the identifier without consuming.
	end := l.pos
	for end < len(l.input) && (isLetter(l.input[end]) || isDigit(l.input[end])) {
		end++
	}
	word := l.input[l.pos:end]
	tt := token.LookupIdent(word)
	return token.StartsStatement(tt)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}
