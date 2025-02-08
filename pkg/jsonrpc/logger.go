package jsonrpc

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "jsonrpc: ", log.LstdFlags)
