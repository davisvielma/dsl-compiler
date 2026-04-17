package parser

import (
	"dsl-compiler/pkg/ast"
	"dsl-compiler/pkg/lexer"
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Item
	peekToken lexer.Item

	entityNames map[string]bool
	routePaths  map[string]bool
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:           l,
		errors:      []string{},
		entityNames: make(map[string]bool),
		routePaths:  make(map[string]bool),
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextItem()
}

func (p *Parser) curTokenIs(t lexer.ItemType) bool {
	return p.curToken.Typ == t
}

func (p *Parser) peekTokenIs(t lexer.ItemType) bool {
	return p.peekToken.Typ == t
}

func (p *Parser) expectPeek(t lexer.ItemType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.ItemType) {
	msg := fmt.Sprintf("Se esperaba el token %s, pero se obtuvo %s", lexer.TokensNames[t], lexer.TokensNames[p.peekToken.Typ])
	p.errors = append(p.errors, msg)
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.ItemEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Typ {
	case lexer.ItemServer:
		return p.parseServerStatement()
	case lexer.ItemEntity:
		return p.parseEntityStatement()
	case lexer.ItemRoute:
		return p.parseRouteStatement()
	default:
		return nil
	}
}

func (p *Parser) parseEntityStatement() *ast.EntityStatement {
	stmt := &ast.EntityStatement{Token: p.curToken}

	if !p.expectPeek(lexer.ItemIdentifier) {
		return nil
	}

	name := p.curToken.Val
	if p.entityNames[strings.ToLower(name)] {
		p.addError(fmt.Sprintf("La entidad '%s' ya ha sido definida", name))
		return nil
	}
	p.entityNames[strings.ToLower(name)] = true

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: name}

	if !p.expectPeek(lexer.ItemLeftBrace) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(lexer.ItemRightBrace) && !p.curTokenIs(lexer.ItemEOF) {
		p.skipComments()

		if p.curTokenIs(lexer.ItemRightBrace) {
			break
		}

		if p.curTokenIs(lexer.ItemIdentifier) {
			field := p.parseFieldDefinition()
			if field != nil {
				stmt.Fields = append(stmt.Fields, field)
			}
		} else {
			msg := fmt.Sprintf(
				"En la entidad '%s': se esperaba el nombre de un campo, pero se encontró el token %s con valor '%s'",
				stmt.Name.Value,
				lexer.TokensNames[p.curToken.Typ],
				p.curToken.Val,
			)
			p.addError(msg)
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFieldDefinition() *ast.FieldDefinition {
	field := &ast.FieldDefinition{Token: p.curToken}
	field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Val}

	if !p.expectPeek(lexer.ItemColon) {
		return nil
	}

	if p.peekTokenIs(lexer.ItemLeftBracket) {
		p.nextToken()
		if !p.expectPeek(lexer.ItemRightBracket) {
			return nil
		}
		field.IsArray = true
	}

	if !p.expectPeek(lexer.ItemIdentifier) {
		return nil
	}
	field.DataType = &ast.Identifier{Token: p.curToken, Value: p.curToken.Val}

	if p.peekTokenIs(lexer.ItemLeftParen) {
		p.parseFieldModifiers(field)
	}

	if p.peekTokenIs(lexer.ItemAssign) {
		p.nextToken()
		p.nextToken()
		field.DefaultValue = p.parseLiteral()
	}

	return field
}

func (p *Parser) parseFieldModifiers(field *ast.FieldDefinition) {
	p.nextToken()

	for !p.curTokenIs(lexer.ItemRightParen) && !p.curTokenIs(lexer.ItemEOF) {
		if p.curTokenIs(lexer.ItemIdentifier) {
			switch strings.ToLower(p.curToken.Val) {
			case "unique":
				field.IsUnique = true
			case "optional":
				field.IsOptional = true
			}
		}
		p.nextToken()
	}
}

func (p *Parser) parseLiteral() ast.Expression {
	switch p.curToken.Typ {
	case lexer.ItemString:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Val}
	case lexer.ItemInt:
		val, _ := strconv.ParseInt(p.curToken.Val, 10, 64)
		return &ast.IntegerLiteral{Token: p.curToken, Value: val}
	case lexer.ItemFloat:
		val, _ := strconv.ParseFloat(p.curToken.Val, 64)
		return &ast.FloatLiteral{Token: p.curToken, Value: val}
	case lexer.ItemBoolean:
		val := p.curToken.Val == "true"
		return &ast.BooleanLiteral{Token: p.curToken, Value: val}
	default:
		return nil
	}
}

func (p *Parser) parseServerStatement() *ast.ServerStatement {
	stmt := &ast.ServerStatement{Token: p.curToken}

	if !p.expectPeek(lexer.ItemIdentifier) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Val}

	if !p.expectPeek(lexer.ItemLeftBrace) {
		return nil
	}

	for !p.curTokenIs(lexer.ItemRightBrace) && !p.curTokenIs(lexer.ItemEOF) {
		if p.curTokenIs(lexer.ItemPort) {
			if !p.expectPeek(lexer.ItemColon) {
				return nil
			}
			if !p.expectPeek(lexer.ItemInt) {
				return nil
			}
			port, _ := strconv.Atoi(p.curToken.Val)
			stmt.Port = port
		} else if p.curTokenIs(lexer.ItemDb) {
			if !p.expectPeek(lexer.ItemColon) {
				return nil
			}
			if !p.expectPeek(lexer.ItemIdentifier) {
				return nil
			}
			stmt.Database = p.curToken.Val
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseRouteStatement() *ast.RouteStatement {
	stmt := &ast.RouteStatement{Token: p.curToken}

	if !p.expectPeek(lexer.ItemString) {
		return nil
	}
	stmt.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Val}

	if !p.expectPeek(lexer.ItemLeftBrace) {
		return nil
	}

	for !p.curTokenIs(lexer.ItemRightBrace) && !p.curTokenIs(lexer.ItemEOF) {
		if p.curTokenIs(lexer.ItemMethods) {
			p.parseRouteMethods(stmt)
		}

		if p.curTokenIs(lexer.ItemTarget) {
			if !p.expectPeek(lexer.ItemColon) {
				return nil
			}
			if !p.expectPeek(lexer.ItemIdentifier) {
				return nil
			}
			stmt.Target = &ast.Identifier{Token: p.curToken, Value: p.curToken.Val}
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseRouteMethods(stmt *ast.RouteStatement) {
	if !p.expectPeek(lexer.ItemColon) {
		return
	}

	p.nextToken()

	for !p.curTokenIs(lexer.ItemRightBrace) && !p.curTokenIs(lexer.ItemTarget) && !p.curTokenIs(lexer.ItemEOF) {
		if p.curTokenIs(lexer.ItemIdentifier) {
			stmt.Methods = append(stmt.Methods, &ast.Identifier{Token: p.curToken, Value: p.curToken.Val})
		}
		if p.peekTokenIs(lexer.ItemComma) {
			p.nextToken()
		}
		p.nextToken()
	}

	if len(stmt.Methods) == 0 {
		msg := fmt.Sprintf(
			"En la ruta '%s' se debe definir al menos un método después de 'METHODS'",
			stmt.Path.Value,
		)
		p.addError(msg)
	}
}

func (p *Parser) skipComments() {
	for p.curToken.Typ == lexer.ItemLineComment || p.curToken.Typ == lexer.ItemBlockComment {
		p.nextToken()
	}
}
