// Package sqllex lexes the SQL sqlc hands the plugin, and the pyformat text
// the MySQL rewriter produces from it, far enough to tell a bindable
// placeholder from a "?" that merely sits inside a string, an identifier or a
// comment. The rewriter and the drivers' argument ordering share it so the
// rules cannot drift apart.
package sqllex

// Dialect is the rule set for one SQL text. Its fields are unexported and
// only the values below are exported: every caller picks a named dialect
// instead of assembling flags, so no call site can lex with rules the
// producer of the text never used.
type Dialect struct {
	// placeholder is one bindable slot as it appears in the text.
	placeholder string
	// escaped is a placeholder lookalike that is a literal instead ("%%" in
	// pyformat text); empty when the dialect has none.
	escaped string
	// numbered marks placeholders that carry a digit suffix (sqlite's ?N).
	numbered bool
	// backslashEscapes marks '...' and "..." as honoring backslash escapes
	// in addition to doubled quotes.
	backslashEscapes bool
	// backtickIdents and bracketIdents mark `...` and [...] as quoted
	// identifiers, whose contents never bind.
	backtickIdents bool
	bracketIdents  bool
	// hashComments marks "#" as a line-comment introducer.
	hashComments bool
	// dashNeedsGap requires whitespace (or end of input) after "--" for it
	// to start a comment; "a--1" is double unary minus.
	dashNeedsGap bool
	// liveVersionComments marks the body of a /*! comment as executable SQL
	// that can hold placeholders.
	liveVersionComments bool
}

// The three texts the plugin lexes. MySQLRaw and MySQLPyformat describe the
// same grammar either side of the rewrite, which is why they must be defined
// together: the rewriter reads the first and every consumer of its output
// reads the second.
var (
	// MySQLRaw is sqlc's MySQL output, before the pyformat rewrite. Only
	// default sql_mode is supported: sqlc's dolphin parser lexes with
	// backslash escapes on and treats "..." as a string, so a query that
	// reached the plugin already parsed under those rules.
	MySQLRaw = Dialect{ //nolint:gochecknoglobals
		placeholder:         "?",
		escaped:             "",
		numbered:            false,
		backslashEscapes:    true,
		backtickIdents:      true,
		bracketIdents:       false,
		hashComments:        true,
		dashNeedsGap:        true,
		liveVersionComments: true,
	}

	// MySQLPyformat is the same text after the rewrite, where placeholders
	// are "%s" and a literal percent has been doubled.
	MySQLPyformat = Dialect{ //nolint:gochecknoglobals
		placeholder:         "%s",
		escaped:             "%%",
		numbered:            false,
		backslashEscapes:    true,
		backtickIdents:      true,
		bracketIdents:       false,
		hashComments:        true,
		dashNeedsGap:        true,
		liveVersionComments: true,
	}

	// SQLite is sqlc's SQLite output, which the drivers execute unchanged.
	// sqlc numbers named parameters (?1, ?2); "#" is not a comment and
	// backslashes are not escapes, but both `...` and [...] quote
	// identifiers.
	SQLite = Dialect{ //nolint:gochecknoglobals
		placeholder:         "?",
		escaped:             "",
		numbered:            true,
		backslashEscapes:    false,
		backtickIdents:      true,
		bracketIdents:       true,
		hashComments:        false,
		dashNeedsGap:        false,
		liveVersionComments: false,
	}
)
