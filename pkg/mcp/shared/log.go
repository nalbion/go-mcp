package shared

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "mcp: ", log.LstdFlags)
