package testmark

import (
	"flag"
)

var Regen = flag.Bool("testmark.regen", false, "Setting this flag hints to systems using testmark tests that fixtures should be regenerated.")
