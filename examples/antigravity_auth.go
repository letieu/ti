package main

import (
	"fmt"
	"log"

	"github.com/letieu/ti/internal/auth"
)

func main() {
	fmt.Println("=== Antigravity Authentication Example ===\n")

	// Check if already authenticated
	if auth.IsAuthenticated(auth.ProviderAntigravity) {
		fmt.Println("✓ Already authenticated with Antigravity")

		// Load existing credentials
		creds, err := auth.LoadCredentials(auth.ProviderAntigravity)
		if err != nil {
			log.Fatalf("Failed to load credentials: %v", err)
		}

		fmt.Printf("Access Token: %s...\n", creds.Access[:20])
		fmt.Printf("Expires At: %d\n", creds.Expires)

		// Display metadata
		if creds.Metadata != nil {
			fmt.Println("\nMetadata:")
			if email, ok := creds.Metadata["email"]; ok {
				fmt.Printf("  Email: %s\n", email)
			}
			if projectId, ok := creds.Metadata["project_id"]; ok {
				fmt.Printf("  Project ID: %s\n", projectId)
			}
		}
		fmt.Println()
		return
	}

	fmt.Println("Starting authentication flow...")
	fmt.Println("A browser window will open for you to authenticate.")
	fmt.Println()

	// Create antigravity auth
	antigravity := auth.AntigravityAuth{}

	// Perform login
	creds, err := antigravity.Login()
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("\n✓ Authentication successful!")
	fmt.Printf("Access Token: %s...\n", creds.Access[:20])
	fmt.Printf("Refresh Token: %s...\n", creds.Refresh[:20])
	fmt.Printf("Expires At: %d\n", creds.Expires)

	// Display metadata
	if creds.Metadata != nil {
		fmt.Println("\nMetadata:")
		if email, ok := creds.Metadata["email"]; ok {
			fmt.Printf("  Email: %s\n", email)
		}
		if projectId, ok := creds.Metadata["project_id"]; ok {
			fmt.Printf("  Project ID: %s\n", projectId)
		}
	}
	fmt.Println()

	// Save credentials
	if err := auth.SaveCredentials(creds, auth.ProviderAntigravity); err != nil {
		log.Fatalf("Failed to save credentials: %v", err)
	}

	fmt.Println("✓ Credentials saved successfully")
	fmt.Println("\nRun this example again to see credentials loaded from storage.")
}
