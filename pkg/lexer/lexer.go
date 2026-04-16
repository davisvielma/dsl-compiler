package lexer

import (
	"strings"
	"unicode/utf8"
)

type stateFn func(*lexer) stateFn

type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan Item
	state stateFn
}

func Lex(input string) *lexer {
	l := &lexer{
		input: input,
		state: lexText,
		items: make(chan Item, 2),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state != nil {
		l.state = l.state(l)
	}
	close(l.items)
}

func (l *lexer) NextItem() Item {
	return <-l.items
}

// --- Herramientas de movimiento ---

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return -1
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

func (l *lexer) ignore() { l.start = l.pos }
func (l *lexer) backup() { l.pos -= l.width }
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) emit(t ItemType) {
	l.items <- Item{Typ: t, Val: l.input[l.start:l.pos]}
	l.start = l.pos
}

// --- Lógica de Estados ---

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], "//") {
			return lexLineComment
		}
		if strings.HasPrefix(l.input[l.pos:], "/*") {
			return lexBlockComment
		}

		r := l.next()
		if r == -1 {
			break
		}
		if isSpace(r) {
			l.ignore()
			continue
		}

		switch r {
		case '{':
			l.emit(ItemLeftBrace)
		case '}':
			l.emit(ItemRightBrace)
		case '(':
			l.emit(ItemLeftParen)
		case ')':
			l.emit(ItemRightParen)
		case '[':
			l.emit(ItemLeftBracket)
		case ']':
			l.emit(ItemRightBracket)
		case '=':
			l.emit(ItemEquals)
		case ':':
			l.emit(ItemColon)
		case ',':
			l.emit(ItemComma)
		case '/':
			l.backup()
			return lexIdentifier
		default:
			if isAlphaNumeric(r) {
				l.backup()
				return lexIdentifier
			}
			l.emit(ItemUnknown)
		}
	}
	l.emit(ItemEOF)
	return nil
}

func lexLineComment(l *lexer) stateFn {
	for {
		r := l.next()
		if r == '\n' || r == -1 {
			break
		}
	}
	l.emit(ItemLineComment)
	return lexText
}

// REQUISITO 4: Comentarios anidados
func lexBlockComment(l *lexer) stateFn {
	l.pos += 2 // Saltar el primer /*
	depth := 1
	for depth > 0 {
		if strings.HasPrefix(l.input[l.pos:], "/*") {
			l.pos += 2
			depth++
			continue
		}
		if strings.HasPrefix(l.input[l.pos:], "*/") {
			l.pos += 2
			depth--
			continue
		}
		if r := l.next(); r == -1 {
			break
		}
	}
	l.emit(ItemBlockComment)
	return lexText
}

func lexIdentifier(l *lexer) stateFn {
	for {
		r := l.next()
		if !isAlphaNumeric(r) && r != '/' {
			l.backup()
			break
		}
	}
	word := strings.ToLower(l.input[l.start:l.pos])
	switch word {
	case "server":
		l.emit(ItemServer)
	case "port":
		l.emit(ItemPort)
	case "db":
		l.emit(ItemDb)
	case "entity":
		l.emit(ItemEntity)
	case "route":
		l.emit(ItemRoute)
	case "methods":
		l.emit(ItemMethods)
	case "target":
		l.emit(ItemTarget)
	default:
		l.emit(ItemIdentifier)
	}
	return lexText
}

func isSpace(r rune) bool { return r == ' ' || r == '\n' || r == '\t' || r == '\r' }
func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
