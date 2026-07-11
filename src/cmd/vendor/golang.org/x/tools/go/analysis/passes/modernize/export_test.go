
// This file exposes yet-unpublished analyzers to the tests.

package modernize

var (
	ImportCommentAnalyzer     = importCommentAnalyzer
	ReflectTypeAssertAnalyzer = reflectTypeAssertAnalyzer
	SlicesBackwardAnalyzer    = slicesBackwardAnalyzer
	UnsafeFuncsAnalyzer       = unsafeFuncsAnalyzer
)
