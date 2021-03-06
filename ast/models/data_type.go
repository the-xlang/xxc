package models

import (
	"strings"
	"unicode"

	"github.com/the-xlang/xxc/lex/tokens"
	"github.com/the-xlang/xxc/pkg/x"
	"github.com/the-xlang/xxc/pkg/xapi"
	"github.com/the-xlang/xxc/pkg/xtype"
)

// DataType is data type identifier.
type DataType struct {
	// Tok used for usually *File comparisons.
	// For this reason, you don't use token as value, identifier or etc.
	Tok             Tok
	Id              uint8
	Original        any
	Kind            string
	MultiTyped      bool
	Tag             any
	DontUseOriginal bool
}

// KindWithOriginalId returns dt.Kind with OriginalId.
func (dt *DataType) KindWithOriginalId() string {
	if dt.Original == nil {
		return dt.Kind
	}
	_, prefix := dt.KindId()
	original := dt.Original.(DataType)
	id, _ := original.KindId()
	return prefix + id
}

// OriginalKindId returns dt.Kind's identifier of official.
//
// Special case is:
//   OriginalKindId() -> "" if DataType has not original
func (dt *DataType) OriginalKindId() string {
	if dt.Original == nil {
		return ""
	}
	t := dt.Original.(DataType)
	id, _ := t.KindId()
	return id
}

// KindId returns dt.Kind's identifier.
func (dt *DataType) KindId() (id, prefix string) {
	if dt.Id == xtype.Map || dt.Id == xtype.Func {
		return dt.Kind, ""
	}
	id = dt.Kind
	runes := []rune(dt.Kind)
	for i, r := range dt.Kind {
		if r == '_' || unicode.IsLetter(r) {
			id = string(runes[i:])
			prefix = string(runes[:i])
			break
		}
	}
	for _, dt := range xtype.TypeMap {
		if dt == id {
			return
		}
	}
	runes = []rune(id)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ':' && i+1 < len(runes) && runes[i+1] == ':' { // Namespace?
			i++
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			id = string(runes[:i])
			break
		}
	}
	return
}

func (dt *DataType) SetToOriginal() {
	if dt.DontUseOriginal || dt.Original == nil {
		return
	}
	tag := dt.Tag
	switch tag.(type) {
	case Genericable:
		defer func() { dt.Tag = tag }()
	}
	kind := dt.KindWithOriginalId()
	tok := dt.Tok
	*dt = dt.Original.(DataType)
	dt.Kind = kind
	dt.Tok = tok
	if strings.HasPrefix(dt.Kind, x.Prefix_Array) {
		dt.Tag = tag
	}
}

// Pointers returns pointer marks of data type.
func (dt *DataType) Pointers() string {
	for i, run := range dt.Kind {
		if run != '*' {
			return dt.Kind[:i]
		}
	}
	return ""
}

func (dt DataType) String() (s string) {
	dt.SetToOriginal()
	if dt.MultiTyped {
		return dt.MultiTypeString()
	}
	// Remove namespace
	i := strings.LastIndex(dt.Kind, tokens.DOUBLE_COLON)
	if i != -1 {
		dt.Kind = dt.Kind[i+len(tokens.DOUBLE_COLON):]
	}
	pointers := dt.Pointers()
	// Apply pointers
	defer func() {
		var cxx strings.Builder
		for range pointers {
			cxx.WriteString("ptr<")
		}
		cxx.WriteString(s)
		for range pointers {
			cxx.WriteString(">")
		}
		s = cxx.String()
	}()
	dt.Kind = dt.Kind[len(pointers):]
	if dt.Kind != "" {
		switch {
		case strings.HasPrefix(dt.Kind, x.Prefix_Slice):
			return dt.SliceString()
		case strings.HasPrefix(dt.Kind, x.Prefix_Array):
			return dt.ArrayString()
		case dt.Id == xtype.Map && dt.Kind[0] == '[' && dt.Kind[len(dt.Kind)-1] == ']':
			return dt.MapString()
		}
	}
	switch dt.Tag.(type) {
	case CompiledStruct:
		return dt.StructString()
	}
	switch dt.Id {
	case xtype.Id, xtype.Enum:
		return xapi.OutId(dt.Kind, dt.Tok.File)
	case xtype.Trait:
		return dt.TraitString()
	case xtype.Struct:
		return dt.StructString()
	case xtype.Func:
		return dt.FuncString()
	default:
		return xtype.CxxId(dt.Id)
	}
}

// SliceString returns cxx value of slice data type.
func (dt DataType) SliceString() string {
	var cxx strings.Builder
	cxx.WriteString("slice<")
	dt.Kind = dt.Kind[len(x.Prefix_Slice):] // Remove slice
	cxx.WriteString(dt.String())
	cxx.WriteByte('>')
	return cxx.String()
}

// ArrayComponent returns data type of array components.
func (dt DataType) ArrayComponent() DataType {
	dt.Kind = dt.Kind[len(x.Prefix_Array):] // Remove array
	exprs := dt.Tag.([][]any)[1:]
	dt.Tag = exprs
	return dt
}

// ArrayString returns cxx value of map data type.
func (dt DataType) ArrayString() string {
	var cxx strings.Builder
	cxx.WriteString("array<")
	exprs := dt.Tag.([][]any)
	expr := exprs[0][1].(Expr)
	cxx.WriteString(dt.ArrayComponent().String())
	cxx.WriteByte(',')
	cxx.WriteString(expr.String())
	cxx.WriteByte('>')
	return cxx.String()
}

// MapString returns cxx value of map data type.
func (dt *DataType) MapString() string {
	var cxx strings.Builder
	types := dt.Tag.([]DataType)
	cxx.WriteString("map<")
	key := types[0]
	key.DontUseOriginal = dt.DontUseOriginal
	cxx.WriteString(key.String())
	cxx.WriteByte(',')
	value := types[1]
	value.DontUseOriginal = dt.DontUseOriginal
	cxx.WriteString(value.String())
	cxx.WriteByte('>')
	return cxx.String()
}

// TraitString returns cxx value of trait data type.
func (dt *DataType) TraitString() string {
	var cxx strings.Builder
	id, _ := dt.KindId()
	cxx.WriteString("trait<")
	cxx.WriteString(xapi.OutId(id, dt.Tok.File))
	cxx.WriteByte('>')
	return cxx.String()
}

// StructString returns cxx value of struct data type.
func (dt *DataType) StructString() string {
	var cxx strings.Builder
	s := dt.Tag.(CompiledStruct)
	cxx.WriteString(s.OutId())
	types := s.Generics()
	if len(types) == 0 {
		return cxx.String()
	}
	cxx.WriteByte('<')
	for _, t := range types {
		t.DontUseOriginal = dt.DontUseOriginal
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">"
}

// FuncString returns cxx value of function DataType.
func (dt *DataType) FuncString() string {
	var cxx strings.Builder
	cxx.WriteString("std::function<")
	f := dt.Tag.(*Func)
	f.RetType.Type.DontUseOriginal = dt.DontUseOriginal
	cxx.WriteString(f.RetType.String())
	cxx.WriteByte('(')
	if len(f.Params) > 0 {
		for _, param := range f.Params {
			param.Type.DontUseOriginal = dt.DontUseOriginal
			cxx.WriteString(param.Prototype())
			cxx.WriteByte(',')
		}
		cxxStr := cxx.String()[:cxx.Len()-1]
		cxx.Reset()
		cxx.WriteString(cxxStr)
	} else {
		cxx.WriteString("void")
	}
	cxx.WriteString(")>")
	return cxx.String()
}

// MultiTypeString returns cxx value of muli-typed DataType.
func (dt *DataType) MultiTypeString() string {
	types := dt.Tag.([]DataType)
	var cxx strings.Builder
	cxx.WriteString("std::tuple<")
	for _, t := range types {
		t.DontUseOriginal = dt.DontUseOriginal
		cxx.WriteString(t.String())
		cxx.WriteByte(',')
	}
	return cxx.String()[:cxx.Len()-1] + ">" + dt.Pointers()
}

// MapKind returns data type kind string of map data type.
func (dt *DataType) MapKind() string {
	types := dt.Tag.([]DataType)
	var kind strings.Builder
	kind.WriteByte('[')
	kind.WriteString(types[0].Kind)
	kind.WriteByte(':')
	kind.WriteString(types[1].Kind)
	kind.WriteByte(']')
	return kind.String()
}
