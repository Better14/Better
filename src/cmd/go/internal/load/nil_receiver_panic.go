// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package load

import (
	"os"
	"strings"

	"cmd/go/internal/modinfo"
	"golang.org/x/mod/modfile"
)

// NilReceiverPanicFromMod returns the nil_receiver_panic mode from a main module's go.mod.
// The empty string means disable (upstream / dependency semantics).
func NilReceiverPanicFromMod(m *modinfo.ModulePublic) string {
	if m == nil || !m.Main {
		return ""
	}
	data, err := os.ReadFile(m.GoMod)
	if err != nil {
		return "enable"
	}
	f, err := modfile.Parse(m.GoMod, data, nil)
	if err != nil || f.NilReceiverPanic == nil {
		if mode := parseNilReceiverPanicLine(string(data)); mode != "" {
			return mode
		}
		return "enable"
	}
	return f.NilReceiverPanic.Mode
}

func parseNilReceiverPanicLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nil_receiver_panic ") {
			f := strings.Fields(line)
			if len(f) == 2 && (f[1] == "enable" || f[1] == "disable") {
				return f[1]
			}
		}
	}
	return ""
}
