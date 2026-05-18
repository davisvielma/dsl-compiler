package analyzer

import (
	"dsl-compiler/pkg/ast"
	"fmt"
	"strings"
)

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
	a.CollectDefinitions()
	a.ValidateReferences()
	a.ValidateDefaultValues()
	return a.errors
}

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

			entitySym := &EntitySymbol{
				Name:   entityName,
				Fields: make(map[string]FieldSymbol),
			}

			for _, field := range s.Fields {
				fieldName := field.Name.Value

				// NUEVA REGLA 1: Bloquear el campo 'id' manual
				if strings.ToLower(fieldName) == "id" {
					a.addError(fmt.Sprintf("Error Semántico en '%s': El campo 'id' es reservado y el compilador lo generará automáticamente.", entityName))
					continue
				}

				if _, fieldExists := entitySym.Fields[fieldName]; fieldExists {
					a.addError(fmt.Sprintf("Error Semántico: Campo '%s' duplicado en la entidad '%s'.", fieldName, entityName))
					continue
				}

				entitySym.Fields[field.Name.Value] = FieldSymbol{
					Name:       fieldName,
					Type:       field.DataType.Value,
					IsArray:    field.IsArray,
					IsOptional: field.IsOptional,
					IsUnique:   field.IsUnique,
					IsRelation: false,
				}
			}

			a.symbolTable.Entities[entityName] = entitySym

		case *ast.RouteStatement:
			path := s.Path.Value

			// ¿Ya existe este path?
			if _, exists := a.symbolTable.Routes[path]; exists {
				a.addError(fmt.Sprintf("Error Semántico: La ruta '%s' ya está definida.", path))
				continue
			}

			routeSym := &RouteSymbol{
				Path:                 path,
				Target:               s.Target.Value,
				HasIndividualActions: false,
			}

			methodCheck := make(map[string]bool)
			for _, m := range s.Methods {
				methodName := strings.ToUpper(m.Value)
				if methodCheck[methodName] {
					a.addError(fmt.Sprintf("Error Semántico: Método '%s' duplicado en la ruta '%s'.", methodName, path))
					continue
				}
				methodCheck[methodName] = true

				if methodName == "GET" || methodName == "PUT" || methodName == "DELETE" {
					routeSym.HasIndividualActions = true
				}

				routeSym.Methods = append(routeSym.Methods, methodName)
			}

			a.symbolTable.Routes[path] = routeSym
		}
	}
}

func (a *SemanticAnalyzer) ValidateReferences() {
	for path, route := range a.symbolTable.Routes {
		if _, exists := a.symbolTable.Entities[route.Target]; !exists {
			a.addError(fmt.Sprintf("Error Semántico en ruta '%s': El TARGET '%s' no existe como ENTITY.", path, route.Target))
		}

		allowedMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true}
		for _, method := range route.Methods {
			if !allowedMethods[method] {
				a.addError(fmt.Sprintf("Error Semántico en ruta '%s': El método '%s' no es válido.", path, method))
			}
		}
	}

	for entityName, entity := range a.symbolTable.Entities {
		for fieldName, field := range entity.Fields {
			if !a.isValidType(field.Type) {
				a.addError(fmt.Sprintf("Error Semántico en '%s.%s': El tipo '%s' no es un tipo primitivo ni una ENTITY definida.", entityName, fieldName, field.Type))
				continue
			}

			if !a.isPrimitive(field.Type) {
				fSym := entity.Fields[fieldName]
				fSym.IsRelation = true
				entity.Fields[fieldName] = fSym
			}
		}
	}
}

func (a *SemanticAnalyzer) ValidateDefaultValues() {
	for _, stmt := range a.program.Statements {
		entity, ok := stmt.(*ast.EntityStatement)
		if !ok {
			continue
		}

		for _, field := range entity.Fields {
			if field.DefaultValue == nil {
				continue
			}

			valueType := a.getLiteralType(field.DefaultValue)

			if field.DataType.Value != valueType {
				a.addError(fmt.Sprintf(
					"Error Semántico en '%s.%s': No se puede asignar un valor de tipo '%s' a un campo de tipo '%s'.",
					entity.Name.Value, field.Name.Value, valueType, field.DataType.Value,
				))
			}
		}
	}
}

// getLiteralType identifica si el nodo es string, int, bool, etc.
func (a *SemanticAnalyzer) getLiteralType(exp ast.Expression) string {
	switch exp.(type) {
	case *ast.IntegerLiteral:
		return "int"
	case *ast.StringLiteral:
		return "string"
	case *ast.BooleanLiteral:
		return "bool"
	case *ast.FloatLiteral:
		return "float"
	default:
		return "unknown"
	}
}

func (a *SemanticAnalyzer) isPrimitive(t string) bool {
	return t == "int" || t == "string" || t == "float" || t == "bool"
}

// isValidType es una función auxiliar para el chequeo de tipos
func (a *SemanticAnalyzer) isValidType(t string) bool {
	if a.isPrimitive(t) {
		return true
	}

	_, isEntity := a.symbolTable.Entities[t]
	return isEntity
}

func (a *SemanticAnalyzer) addError(msg string) {
	a.errors = append(a.errors, msg)
}

func (a *SemanticAnalyzer) Errors() []string {
	return a.errors
}

func (a *SemanticAnalyzer) GetSymbolTable() *SymbolTable {
	return a.symbolTable
}
