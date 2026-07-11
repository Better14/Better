// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

import "go/token"

// AsTypeSpec returns an equivalent type specification for a struct shorthand decl.
func (s *StructDecl) AsTypeSpec() *TypeSpec {
	return &TypeSpec{
		Name:       s.Name,
		TypeParams: s.TypeParams,
		Type: &StructType{
			Struct: s.Struct,
			Fields: s.Fields,
		},
	}
}

// AsTypeSpec returns an equivalent type specification for an interface shorthand decl.
func (s *InterfaceDecl) AsTypeSpec() *TypeSpec {
	return &TypeSpec{
		Name:       s.Name,
		TypeParams: s.TypeParams,
		Type: &InterfaceType{
			Interface: s.Interface,
			Methods:   s.Methods,
		},
	}
}

// ShorthandKeywordPos returns the keyword position for shorthand type declarations.
func ShorthandKeywordPos(d Decl) token.Pos {
	switch d := d.(type) {
	case *StructDecl:
		return d.Struct
	case *InterfaceDecl:
		return d.Interface
	default:
		return token.NoPos
	}
}
