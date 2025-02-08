package mcp

// import (
// 	"context"
// 	"log"
// 	"os"
// 	"strconv"

// 	"github.com/nalbion/go-mcp/pkg/mcp/client"
// )

// func Main() {
// 	args := os.Args[1:]
// 	if len(args) < 1 {
// 		log.Fatalf("Usage: %s <client|server> [args...]", os.Args[0])
// 	}

// 	ctx := context.Background()
// 	command := args[0]
// 	switch command {
// 	case "client":
// 		if len(args) < 2 {
// 			log.Fatalf("Usage: client <server_url_or_command> [args...]")
// 		}
// 		client.RunClient(ctx, args[1], args[2:])
// 	case "server":
// 		port := 8080
// 		if len(args) > 1 {
// 			var err error
// 			port, err = strconv.Atoi(args[1])
// 			if err != nil {
// 				log.Fatalf("Invalid port: %v", err)
// 			}
// 		}
// 		runServer(port)
// 	default:
// 		log.Fatalf("Unrecognized command: %s", command)
// 	}
// }
