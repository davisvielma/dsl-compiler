package ast

import (
	"bytes"
	"dsl-compiler/pkg/lexer"
	"fmt"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Statements
type ServerStatement struct {
	Token    lexer.Item
	Name     *Identifier
	Port     int
	Database string
}

func (ss *ServerStatement) statementNode()       {}
func (ss *ServerStatement) TokenLiteral() string { return ss.Token.Val }
func (ss *ServerStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ss.TokenLiteral() + " ")
	out.WriteString(ss.Name.String() + " {\n")
	out.WriteString(fmt.Sprintf("  PORT: %d\n", ss.Port))
	out.WriteString("  DB: " + ss.Database + "\n")
	out.WriteString("}\n")
	return out.String()
}

type FieldDefinition struct {
	Token        lexer.Item
	Name         *Identifier
	DataType     *Identifier
	IsArray      bool
	IsUnique     bool
	IsOptional   bool
	DefaultValue Expression
}

func (f *FieldDefinition) String() string {
	var out bytes.Buffer
	out.WriteString("  " + f.Name.String() + ": ")
	if f.IsArray {
		out.WriteString("[]")
	}
	out.WriteString(f.DataType.String())

	// Mostrar modificadores si existen
	if f.IsUnique || f.IsOptional {
		out.WriteString(" (")
		var mods []string
		if f.IsUnique {
			mods = append(mods, "unique")
		}
		if f.IsOptional {
			mods = append(mods, "optional")
		}
		out.WriteString(strings.Join(mods, ", "))
		out.WriteString(")")
	}
	if f.DefaultValue != nil {
		out.WriteString(" = " + f.DefaultValue.String())
	}
	return out.String()
}

type EntityStatement struct {
	Token  lexer.Item
	Name   *Identifier
	Fields []*FieldDefinition
}

func (es *EntityStatement) statementNode()       {}
func (es *EntityStatement) TokenLiteral() string { return es.Token.Val }
func (es *EntityStatement) String() string {
	var out bytes.Buffer
	out.WriteString(es.TokenLiteral() + " ")
	out.WriteString(es.Name.String() + " {\n")
	for _, f := range es.Fields {
		out.WriteString(f.String() + "\n")
	}
	out.WriteString("}\n")
	return out.String()
}

type RouteStatement struct {
	Token   lexer.Item
	Path    *StringLiteral
	Methods []*Identifier
	Target  *Identifier
}

func (rs *RouteStatement) statementNode()       {}
func (rs *RouteStatement) TokenLiteral() string { return rs.Token.Val }
func (rs *RouteStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + rs.Path.String() + " {\n")

	var methods []string
	for _, m := range rs.Methods {
		methods = append(methods, m.String())
	}

	out.WriteString("  METHODS: " + strings.Join(methods, ", ") + "\n")
	out.WriteString("  TARGET: " + rs.Target.String() + "\n")
	out.WriteString("}\n")
	return out.String()
}

type Identifier struct {
	Token lexer.Item
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Val }
func (i *Identifier) String() string       { return i.Value }

type StringLiteral struct {
	Token lexer.Item
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Val }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

type IntegerLiteral struct {
	Token lexer.Item
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Val }
func (il *IntegerLiteral) String() string       { return il.Token.Val }

type FloatLiteral struct {
	Token lexer.Item
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Val }
func (fl *FloatLiteral) String() string       { return fl.Token.Val }

type BooleanLiteral struct {
	Token lexer.Item
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Val }
func (bl *BooleanLiteral) String() string       { return bl.Token.Val }
