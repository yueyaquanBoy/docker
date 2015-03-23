// +build windows

package parser

import "regexp"

// We use a backtick on Windows to avoid issues in Dockerfiles such as
// "WORKDIR \". Otherwise we would have no way of determinically working out
// if that is the root of the image or a continuation character (\ on Linux).

var TOKEN_LINE_CONTINUATION = regexp.MustCompile("`" + `[ \t]*$`)
