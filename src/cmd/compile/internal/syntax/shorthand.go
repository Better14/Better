// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

// AsTypeDecl returns an equivalent type declaration for a struct shorthand decl.
func (s *StructDecl) AsTypeDecl() *TypeDecl {
	st := new(StructType)
	st.pos = s.pos
	st.FieldList = s.FieldList
	st.TagList = s.TagList
	return &TypeDecl{
		Group:      s.Group,
		Pragma:     s.Pragma,
		Name:       s.Name,
		TParamList: s.TParamList,
		Type:       st,
	}
}

// AsTypeDecl returns an equivalent type declaration for an interface shorthand decl.
func (s *InterfaceDecl) AsTypeDecl() *TypeDecl {
	it := new(InterfaceType)
	it.pos = s.pos
	it.MethodList = s.MethodList
	return &TypeDecl{
		Group:      s.Group,
		Pragma:     s.Pragma,
		Name:       s.Name,
		TParamList: s.TParamList,
		Type:       it,
	}
}
