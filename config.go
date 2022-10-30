package boorumux

import (
	"fmt"
)

// Version numbers
const (
	verMajor = 0
	verMinor = 0
	verPatch = 1
)

// Default agents
var (
	userAgent    = fmt.Sprintf("boorumux/%d.%d.%d", verMajor, verMinor, verPatch)
	serverHeader = fmt.Sprintf("boorumux/%d.%d.%d", verMajor, verMinor, verPatch)
)

var verString = fmt.Sprintf("v%d.%d.%d", verMajor, verMinor, verPatch)
