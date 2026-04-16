package lexer

import "fmt"

type ItemType int

const (
	ItemError ItemType = iota
	ItemEOF
	ItemEntity
	ItemRoute
	ItemServer
	ItemPort
	ItemDb
	ItemMethods
	ItemTarget
	ItemIdentifier
	ItemLeftBrace
	ItemRightBrace
	ItemLeftParen
	ItemRightParen
	ItemLeftBracket
	ItemRightBracket
	ItemAssign
	ItemColon
	ItemComma
	ItemSlash
	ItemLineComment
	ItemBlockComment
	ItemString
	ItemInt
	ItemFloat
	ItemBoolean
	ItemUnknown
)

var TokensNames = []string{
	"ERROR", "EOF", "ENTITY", "ROUTE", "SERVER", "PORT", "DB",
	"METHODS", "TARGET", "ID", "LBRACE", "RBRACE", "LPAREN", "RPAREN",
	"LBRACKET", "RBRACKET", "ASSIGN", "COLON", "COMMA", "SLASH",
	"COMMENT_LINE", "COMMENT_BLOCK", "STRING", "INT", "FLOAT", "BOOLEAN", "UNKNOWN",
}

type Item struct {
	Typ ItemType
	Val string
}

func (i Item) String() string {
	if i.Typ == ItemEOF {
		return "EOF"
	}
	return fmt.Sprintf("%q", i.Val)
}
