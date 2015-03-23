// +build !windows

package parser

import "regexp"

var TOKEN_LINE_CONTINUATION = regexp.MustCompile(`\\[ \t]*$`)
