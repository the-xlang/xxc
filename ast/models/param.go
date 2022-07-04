package models

import (
	"strings"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/xapi"
)

// Param is function parameter AST model.
type Param struct {
	Tok       Tok
	Id        string
	Const     bool
	Volatile  bool
	Variadic  bool
	Reference bool
	Type      DataType
	Default   Expr
}

// TypeString returns data type string of parameter.
func (p *Param) TypeString() string {
	var ts strings.Builder
	if p.Variadic {
		ts.WriteString(tokens.TRIPLE_DOT)
	}
	if p.Reference {
		ts.WriteString(tokens.AMPER)
	}
	ts.WriteString(p.Type.Val)
	return ts.String()
}

func (p Param) String() string {
	var cxx strings.Builder
	cxx.WriteString(p.Prototype())
	if p.Id != "" {
		cxx.WriteByte(' ')
		cxx.WriteString(xapi.OutId(p.Id, p.Tok.File))
	}
	return cxx.String()
}

// Prototype returns prototype cxx of parameter.
func (p *Param) Prototype() string {
	var cxx strings.Builder
	if p.Volatile {
		cxx.WriteString("volatile ")
	}
	if p.Const {
		cxx.WriteString("const ")
	}
	if p.Variadic {
		cxx.WriteString("array<")
		cxx.WriteString(p.Type.String())
		cxx.WriteByte('>')
	} else {
		cxx.WriteString(p.Type.String())
	}
	if p.Reference {
		cxx.WriteByte('&')
	}
	return cxx.String()
}