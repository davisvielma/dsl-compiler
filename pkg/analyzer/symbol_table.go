package analyzer

import (
	"dsl-compiler/pkg/ast"
	"fmt"
)

// FieldSymbol guarda los atributos de un campo según tus láminas
type FieldSymbol struct {
	Name       string
	Type       string // Puede ser primitivo (int, string...) o el nombre de otra ENTITY
	IsArray    bool
	IsOptional bool
	IsUnique   bool
}

// EntitySymbol guarda la metadata de la entidad para validar relaciones y campos
type EntitySymbol struct {
	Name   string
	Fields map[string]FieldSymbol
}

// RouteSymbol define la información de una ruta para validar
type RouteSymbol struct {
	Path    string
	Methods []string // Para validar que no haya métodos repetidos
	Target  string   // El nombre de la entidad a la que apunta
}

// SymbolTable es el "cerebro" que mencionamos
type SymbolTable struct {
	Entities map[string]*EntitySymbol
	Routes   map[string]*RouteSymbol
	Server   *ast.ServerStatement
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Entities: make(map[string]*EntitySymbol),
		Routes:   make(map[string]*RouteSymbol),
	}
}

// SemanticAnalyzer manejará las fases de validación
type SemanticAnalyzer struct {
	program     *ast.Program
	symbolTable *SymbolTable
	errors      []string
}

func New(program *ast.Program) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		program:     program,
		symbolTable: NewSymbolTable(),
		errors:      []string{},
	}
}

// Analyze es la función principal que orquesta las fases
func (a *SemanticAnalyzer) Analyze() []string {
	a.CollectDefinitions() // Paso 1: Ya lo hicimos (llenar la tabla)
	a.ValidateReferences() // Paso 2: Lo que haremos ahora
	return a.errors
}

// PASO 1: Recolectar definiciones globales
func (a *SemanticAnalyzer) CollectDefinitions() {
	for _, stmt := range a.program.Statements {
		switch s := stmt.(type) {
		case *ast.ServerStatement:
			if a.symbolTable.Server != nil {
				a.addError("Error Semántico: No puedes definir más de un SERVER.")
			}
			a.symbolTable.Server = s

		case *ast.EntityStatement:
			entityName := s.Name.Value
			if _, exists := a.symbolTable.Entities[entityName]; exists {
				a.addError(fmt.Sprintf("Error Semántico: La entidad '%s' ya ha sido definida.", entityName))
				continue
			}

			// Creamos el símbolo de la entidad
			entitySym := &EntitySymbol{
				Name:   entityName,
				Fields: make(map[string]FieldSymbol),
			}

			// Registro de campos para validar unicidad interna
			for _, field := range s.Fields {
				if _, fieldExists := entitySym.Fields[field.Name.Value]; fieldExists {
					a.addError(fmt.Sprintf("Error Semántico: Campo '%s' duplicado en la entidad '%s'.", field.Name.Value, entityName))
					continue
				}

				entitySym.Fields[field.Name.Value] = FieldSymbol{
					Name:       field.Name.Value,
					Type:       field.DataType.Value,
					IsArray:    field.IsArray,
					IsOptional: field.IsOptional,
					IsUnique:   field.IsUnique,
				}
			}

			a.symbolTable.Entities[entityName] = entitySym

		case *ast.RouteStatement:
			path := s.Path.Value

			// VALIDACIÓN: ¿Ya existe este path?
			if _, exists := a.symbolTable.Routes[path]; exists {
				a.addError(fmt.Sprintf("Error Semántico: La ruta '%s' ya está definida.", path))
				continue
			}

			// Guardamos la ruta en la tabla para tener el registro global
			routeSym := &RouteSymbol{
				Path:   path,
				Target: s.Target.Value,
			}

			// Validamos métodos duplicados dentro de la misma ruta (ej: GET, GET)
			methodCheck := make(map[string]bool)
			for _, m := range s.Methods {
				if methodCheck[m.Value] {
					a.addError(fmt.Sprintf("Error Semántico: Método '%s' duplicado en la ruta '%s'.", m.Value, path))
				}
				methodCheck[m.Value] = true
				routeSym.Methods = append(routeSym.Methods, m.Value)
			}

			a.symbolTable.Routes[path] = routeSym
		}
	}
}

// PASO 2: ValidateReferences verifica que las relaciones y targets tengan sentido
func (a *SemanticAnalyzer) ValidateReferences() {
	// 1. Validar Targets de las Rutas
	for path, route := range a.symbolTable.Routes {
		if _, exists := a.symbolTable.Entities[route.Target]; !exists {
			a.addError(fmt.Sprintf("Error Semántico en ruta '%s': El TARGET '%s' no existe como ENTITY.", path, route.Target))
		}

		// Validar que los métodos sean los permitidos por nuestro estándar
		allowedMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true}
		for _, method := range route.Methods {
			if !allowedMethods[method] {
				a.addError(fmt.Sprintf("Error Semántico en ruta '%s': El método '%s' no es válido.", path, method))
			}
		}
	}

	// 2. Validar tipos de campos (Relaciones y Primitivos)
	for entityName, entity := range a.symbolTable.Entities {
		for fieldName, field := range entity.Fields {
			if !a.isValidType(field.Type) {
				a.addError(fmt.Sprintf("Error Semántico en '%s.%s': El tipo '%s' no es un tipo primitivo ni una ENTITY definida.", entityName, fieldName, field.Type))
			}
		}
	}
}

// isValidType es una función auxiliar para el chequeo de tipos
func (a *SemanticAnalyzer) isValidType(t string) bool {
	// Tipos primitivos básicos que soporta tu compilador
	primitives := map[string]bool{
		"int":    true,
		"string": true,
		"float":  true,
		"bool":   true,
	}

	if primitives[t] {
		return true
	}

	// Si no es primitivo, verificamos si es una relación con otra entidad
	_, isEntity := a.symbolTable.Entities[t]
	return isEntity
}

func (a *SemanticAnalyzer) addError(msg string) {
	a.errors = append(a.errors, msg)
}

func (a *SemanticAnalyzer) Errors() []string {
	return a.errors
}
