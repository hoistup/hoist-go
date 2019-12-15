// Package erks provides error kinds for the hoist-go package.
package erks

import "github.com/JosiahWitt/erk"

// Default error kind.
// Use this instead of erk.DefaultKind.
type Default struct{ erk.DefaultKind }
