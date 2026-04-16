package main

import (
	"dsl-compiler/pkg/lexer"
	"dsl-compiler/pkg/parser"
	"fmt"
	"os"
)

func main() {
	content, err := os.ReadFile("api.dsl")
	if err != nil {
		fmt.Println("Error: No se encontró el archivo api.dsl")
		return
	}

	l := lexer.Lex(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("Error: %s\n", msg)
		}
		return
	}

	fmt.Println("AST Generado correctamente:")
	fmt.Println(program.String())
}
