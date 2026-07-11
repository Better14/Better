
package types

import (
	"go/ast"
	"go/token"
)

func astNewIdent(pos token.Pos, name string) *ast.Ident {
	return &ast.Ident{NamePos: pos, Name: name}
}

func fieldNameIdent(f *ast.Field) *ast.Ident {
	if f == nil || len(f.Names) == 0 {
		return nil
	}
	return f.Names[0]
}

func fieldName(f *ast.Field) string {
	if id := fieldNameIdent(f); id != nil {
		return id.Name
	}
	return ""
}

func enumVariantFields(spec *ast.EnumVariantSpec) []*ast.Field {
	if spec == nil || spec.StructFields == nil {
		return nil
	}
	return spec.StructFields.List
}
