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
	ItemEquals
	ItemColon
	ItemComma
	ItemSlash
	ItemLineComment
	ItemBlockComment
	ItemUnknown
)

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
