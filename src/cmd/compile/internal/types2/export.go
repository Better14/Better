
package types2

import (
	"cmp"
	"slices"
	"strings"
)

// isOperatorName reports whether name is an overloadable operator identifier.
func isOperatorName(name string) bool {
	switch name {
	case "+", "-", "*", "/", "%",
		"&", "|", "^", "&^", "<<", ">>",
		"==", "!=", "<", ">", "<=", ">=",
		"!", "[]", "[]=", "++", "--":
		return true
	}
	return false
}

// ExportNameVisible reports whether an object name should appear in export data
// for importers even when it is not a capitalized Go export name. Operator
// overloads use operator tokens as names and must be visible across packages.
func ExportNameVisible(name string) bool {
	if isExported(name) {
		return true
	}
	base := name
	if i := strings.Index(name, "·"); i >= 0 {
		base = name[:i]
	}
	return isOperatorName(base)
}

func exportObjectName(obj Object) string {
	if fn, ok := obj.(*Func); ok {
		return fn.LinkName()
	}
	return obj.Name()
}

// PackageExportObjects returns the package-scope objects written to export data.
// All overload variants are included, not only the representative in scope.
func PackageExportObjects(pkg *Package) []Object {
	if pkg == nil || pkg.scope == nil {
		return nil
	}
	seen := make(map[Object]bool)
	var objs []Object
	add := func(obj Object) {
		if obj == nil || seen[obj] {
			return
		}
		seen[obj] = true
		objs = append(objs, obj)
	}
	for _, name := range pkg.scope.Names() {
		add(pkg.scope.Lookup(name))
	}
	if pkg.overloadFuncs != nil {
		for _, funcs := range pkg.overloadFuncs {
			for _, fn := range funcs {
				add(fn)
			}
		}
	}
	slices.SortFunc(objs, func(a, b Object) int {
		return cmp.Compare(exportObjectName(a), exportObjectName(b))
	})
	return objs
}
