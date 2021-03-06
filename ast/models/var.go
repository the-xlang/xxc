package models

import (
	"strings"

	"github.com/the-xlang/xxc/pkg/xapi"
)

// Var is variable declaration AST model.
type Var struct {
	Pub       bool
	DefTok    Tok
	IdTok     Tok
	SetterTok Tok
	Id        string
	Type      DataType
	Expr      Expr
	Const     bool
	New       bool
	Tag       any
	ExprTag   any
	Desc      string
	Used      bool
}

// OutId returns xapi.OutId result of var.
func (v *Var) OutId() string {
	return xapi.OutId(v.Id, v.IdTok.File)
}

func (v Var) String() string {
	if v.Const {
		return ""
	}
	var cxx strings.Builder
	cxx.WriteString(v.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(v.OutId())
	expr := v.Expr.String()
	if expr != "" {
		cxx.WriteString(" = ")
		cxx.WriteString(v.Expr.String())
	} else {
		cxx.WriteString(xapi.DefaultExpr)
	}
	cxx.WriteByte(';')
	return cxx.String()
}

// FieldString returns variable as cxx struct field.
func (v *Var) FieldString() string {
	var cxx strings.Builder
	if v.Const {
		cxx.WriteString("const ")
	}
	cxx.WriteString(v.Type.String())
	cxx.WriteByte(' ')
	cxx.WriteString(v.OutId())
	cxx.WriteString(xapi.DefaultExpr)
	cxx.WriteByte(';')
	return cxx.String()
}
