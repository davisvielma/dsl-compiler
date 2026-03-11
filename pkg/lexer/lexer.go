package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type stateFn func(*lexer) stateFn

type lexer struct {
	input string
	start int
	pos   int
	width int
	line  int
	items chan Item
	state stateFn
}

func Lex(input string) *lexer {
	l := &lexer{
		input: input,
		state: lexText,
		line:  1,
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

func (l *lexer) countLines() int {
	return 1 + strings.Count(l.input[:l.start], "\n")
}

func (l *lexer) ignore() { l.start = l.pos }
func (l *lexer) backup() { l.pos -= l.width }
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) emit(t ItemType) {
	l.items <- Item{
		Typ:  t,
		Val:  l.input[l.start:l.pos],
		Line: l.countLines(), // Calculamos la línea real en el punto de inicio del token
	}
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{
		Typ:  ItemError,
		Val:  fmt.Sprintf(format, args...),
		Line: l.countLines(),
	}
	return nil
}

// --- Lógica de Estados (Lexical Grammar) ---

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
		case ':':
			l.emit(ItemColon)
		case ',':
			l.emit(ItemComma)
		case '/':
			// Si es el inicio de una ruta, la tratamos como identificador especial
			l.backup()
			return lexIdentifier
		default:
			if isAlphaNumeric(r) {
				l.backup()
				return lexIdentifier
			}
			return l.errorf("carácter inesperado: %q", r)
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
	l.ignore()
	return lexText
}

func lexBlockComment(l *lexer) stateFn {
	l.pos += 2 // saltar /*
	for {
		if strings.HasPrefix(l.input[l.pos:], "*/") {
			l.pos += 2
			l.ignore()
			return lexText
		}
		if l.next() == -1 {
			return l.errorf("comentario multilínea sin cerrar")
		}
	}
}

func lexIdentifier(l *lexer) stateFn {
	for {
		r := l.next()
		if !isAlphaNumeric(r) && r != '/' {
			l.backup()
			break
		}
	}

	word := l.input[l.start:l.pos]
	loweredWord := strings.ToLower(word)

	switch loweredWord {
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
