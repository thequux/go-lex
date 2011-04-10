/*
Golex is a lexer generator for Go.
It is written in Go and generates lexers in Go.

Input files should be syntactically valid Go source that somehow
imports the pseudo-package "golex", which exports the definitions
below.

Lexers are created by writing a switch statement of the form
 
	switch tok := golex.NextToken(reader); tok {
	// productions
	}

Each case should be a regex (or a list of regexes), formatted as
strings.  The lexer consumes characters from the input until no
pattern matches, and then executes the first case clause that
matched.  If the return value from golex.NextToken is assigned to a
variable, that variable behaves like a golex.Token object.

*/
package documentation

import (
	"bufio"
)
/*
  Of particular note is the Token function, which can only be
used in the expression or initial statement of a switch block; such a
switch block becomes a lexer.
*/

type LexerState struct {
	unexported fields
}

// Initialize a new lexer state, reading from the given stream
func InitLexer(io.Reader) *LexerState

// Retrieve the next token from the input stream. Only valid at the
// top of a switch statement; see package documentation for details
func (*LexerState) NextToken() Token

type Position struct {
	Line, Column int
}

// Lexers return tokens, which behave somewhat like the following structure.
//
// Note that the fields of this structure should be considered macros. 
type Token struct {
	// The contents of the token
	Text string

	// starting and ending positions of the token
	Start, End Position
}