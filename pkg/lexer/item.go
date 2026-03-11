package lexer

import "fmt"

type ItemType int

const (
	ItemError ItemType = iota
	ItemEOF
	ItemEntity     // "entity"
	ItemRoute      // "route"
	ItemServer     // "server"
	ItemPort       // "port"
	ItemDb         // "db"
	ItemMethods    // "methods"
	ItemTarget     // "target"
	ItemIdentifier // Nombres de campos, tipos (int, string), nombres de entidades
	ItemLeftBrace  // "{"
	ItemRightBrace // "}"
	ItemColon      // ":"
	ItemComma      // ","
	ItemSlash      // "/"
)

type Item struct {
	Typ  ItemType
	Val  string
	Line int
}

func (i Item) String() string {
	switch i.Typ {
	case ItemEOF:
		return "EOF"
	case ItemError:
		return fmt.Sprintf("Línea %d: ERROR: %s", i.Line, i.Val)
	}
	return fmt.Sprintf("%q", i.Val)
}
