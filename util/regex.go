package util

import "regexp"

var WordRegex = regexp.MustCompile(`[\w'-]+\S*`)
