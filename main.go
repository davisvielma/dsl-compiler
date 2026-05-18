package main

import (
	"dsl-compiler/pkg/analyzer"
	"dsl-compiler/pkg/generator"
	"dsl-compiler/pkg/lexer"
	"dsl-compiler/pkg/parser"
	"fmt"
	"os"
)

func main() {
	content, err := os.ReadFile("api.dsl")
	if err != nil {
		fmt.Printf("Error crítico: No se pudo leer el archivo 'api.dsl': %v\n", err)
		return
	}

	fmt.Println("--- Iniciando Compilación ---")

	l := lexer.Lex(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("\n❌ Se encontraron %d errores sintácticos:\n", len(p.Errors()))
		for _, msg := range p.Errors() {
			fmt.Printf("  [Sintaxis]: %s\n", msg)
		}
		return
	}
	fmt.Println("✅ Análisis sintáctico completado sin errores.")

	semAnalyzer := analyzer.New(program)
	semanticErrors := semAnalyzer.Analyze()

	if len(semanticErrors) > 0 {
		fmt.Printf("\n❌ Se encontraron %d errores semánticos:\n", len(semanticErrors))
		for _, msg := range semanticErrors {
			fmt.Printf("  [Semántica]: %s\n", msg)
		}
		return
	}
	fmt.Println("✅ Análisis semántico completado sin errores.")

	fmt.Println("\n🔨 Iniciando generación de archivos...")

	gen := generator.New(semAnalyzer.GetSymbolTable())
	err = gen.Generate()
	if err != nil {
		fmt.Printf("\n❌ Error crítico durante la generación: %v\n", err)
		return
	}

	fmt.Println("\n---------------------------------")
	fmt.Println("🚀 Proceso finalizado con éxito.")
}
