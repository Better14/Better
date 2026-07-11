
package types

const linqPkgPath = "linq"

// linqSliceFastPaths maps exported linq extension names to unexported
// []T implementations. The compiler rewrites slice receivers on the
// first call in a chain to these functions instead of wrapping with
// slices.Values.
var linqSliceFastPaths = map[string]string{
	"Where":                  "whereSlice",
	"Select":                 "selectBySlice",
	"SelectBy":               "selectBySlice",
	"Take":                   "takeSlice",
	"Skip":                   "skipSlice",
	"Reverse":                "reverseSlice",
	"Distinct":               "distinctSlice",
	"Contains":               "containsSlice",
	"First":                  "firstFromSlice",
	"FirstOrDefault":         "firstOrDefaultSlice",
	"Last":                   "lastFromSlice",
	"LastOrDefault":          "lastOrDefaultSlice",
	"Single":                 "singleFromSlice",
	"SingleOrDefault":        "singleOrDefaultFromSlice",
	"ElementAt":              "elementAtFromSlice",
	"ElementAtOrDefault":     "elementAtOrDefaultFromSlice",
	"Any":                    "anySlice",
	"All":                    "allSlice",
	"Count":                  "CountSlice",
	"LongCount":              "LongCountSlice",
	"Sum":                    "sumSlice",
	"Average":                "averageSlice",
	"Max":                    "maxSlice",
	"Min":                    "minSlice",
	"MaxBy":                  "maxBySlice",
	"MinBy":                  "minBySlice",
	"Aggregate":              "aggregateSlice",
	"AggregateWithSeed":      "aggregateWithSeedSlice",
	"ToList":                 "toListSlice",
	"ToArray":                "toListSlice",
	"ToHashSet":              "toHashSetSlice",
	"TryGetSeqLen":           "TryGetSeqLenSlice",
}

func linqSliceFastPath(method string) (string, bool) {
	name, ok := linqSliceFastPaths[method]
	return name, ok
}

func isLinqCompilerSliceFast(pkgPath, name string) bool {
	if pkgPath != linqPkgPath {
		return false
	}
	for _, fast := range linqSliceFastPaths {
		if fast == name {
			return true
		}
	}
	switch name {
	case "countSlice", "longCountSlice", "tryGetSeqLenSlice", "countMap", "longCountMap", "tryGetSeqLenMap":
		return true
	}
	return false
}
