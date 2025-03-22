package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/open-xyz/vgx/pkg/api"
	"github.com/open-xyz/vgx/pkg/cache"
)

func main() {
    // Load environment variables
    godotenv.Load()
    
    // Initialize cache
    cacheDir := "/app/cache"
    if err := cache.Initialize(cacheDir); err != nil {
        log.Fatalf("Failed to initialize cache: %v", err)
    }
    
    // Set up HTTP server
    http.HandleFunc("/scan", api.HandleScan)
    http.HandleFunc("/health", api.HandleHealth)
    
    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    fmt.Printf("🚀 VGX service running on port %s\n", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
