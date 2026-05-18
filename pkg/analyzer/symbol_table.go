package analyzer

import (
	"dsl-compiler/pkg/ast"
)

// FieldSymbol guarda los atributos de un campo
type FieldSymbol struct {
	Name         string
	Type         string
	IsArray      bool
	IsOptional   bool
	IsUnique     bool
	IsRelation   bool
	DefaultValue ast.Expression
}

// EntitySymbol guarda la metadata de la entidad para validar relaciones y campos
type EntitySymbol struct {
	Name   string
	Fields map[string]FieldSymbol
}

// RouteSymbol define la información de una ruta para validar
type RouteSymbol struct {
	Path                 string
	Methods              []string
	Target               string
	HasIndividualActions bool // Indica si generará rutas con /:id automáticamente
}

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
