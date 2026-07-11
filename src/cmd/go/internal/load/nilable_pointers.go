
package load

import (
	"os"
	"strings"

	"cmd/go/internal/modinfo"
	"golang.org/x/mod/modfile"
)

// NilablePointersFromMod returns the nilable_pointers mode from a main module's go.mod.
func NilablePointersFromMod(m *modinfo.ModulePublic) string {
	if m == nil || !m.Main || m.GoMod == "" {
		return ""
	}
	data, err := os.ReadFile(m.GoMod)
	if err != nil {
		return ""
	}
	f, err := modfile.Parse(m.GoMod, data, nil)
	if err != nil || f.NilablePointers == nil {
		return parseNilablePointersLine(string(data))
	}
	return f.NilablePointers.Mode
}

// parseNilablePointersLine is a fallback when modfile does not know nilable_pointers.
func parseNilablePointersLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nilable_pointers ") {
			f := strings.Fields(line)
			if len(f) == 2 && (f[1] == "enable" || f[1] == "disable" || f[1] == "warnings") {
				return f[1]
			}
		}
	}
	return ""
}
