package mcp

// func GetSASTFindingsDetailed(service *web.Service, cfg *config.Config) {
// 	u := usecase.NewInteractor(func(ctx context.Context, input *any, output *any) error {
// 		return nil
// 	})

// 	u.SetTitle("MCP server endpoint")
// 	u.SetDescription("Refer to https://modelcontextprotocol.io/introduction")
// 	// u.SetExpectedErrors(status.InvalidArgument, status.PermissionDenied, status.Internal)

// 	service.Get("/sse", u)
// }

// func runServer(port int) {
// 	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		fmt.Println("Got new SSE connection")
// 		// Implement SSE connection handling here
// 	})

// 	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("Received message")
// 		// Implement message handling here
// 	})

// 	log.Printf("Server running on http://localhost:%d/sse", port)
// 	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
// }
