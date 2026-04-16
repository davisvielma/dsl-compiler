package main

import (
	"dsl-compiler/pkg/lexer"
	"fmt"
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("api.dsl")
	if err != nil {
		fmt.Println("Error: No se encontró el archivo api.dsl")
		return
	}

	l := lexer.Lex(string(content))

	fmt.Printf("%-15s | %-20s\n", "TIPO DE TOKEN", "VALOR")
	fmt.Println(strings.Repeat("-", 40))

	for {
		item := l.NextItem()

		fmt.Printf("%-15s | %v\n", formatType(item.Typ), item)

		if item.Typ == lexer.ItemEOF {
			break
		}
	}
}

func formatType(t lexer.ItemType) string {
	names := []string{
		"ERROR", "EOF", "ENTITY", "ROUTE", "SERVER", "PORT", "DB",
		"METHODS", "TARGET", "ID", "LBRACE", "RBRACE", "LPAREN", "RPAREN",
		"LBRACKET", "RBRACKET", "EQUALS", "COLON", "COMMA", "SLASH", "COMMENT_LINE",
		"COMMENT_BLOCK", "UNKNOWN",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "UNKNOWN"
}
