package main

import (
	"fmt"
	"log"
	"time"

	"github.com/letieu/ti/internal/authstore"
	"github.com/letieu/ti/internal/config"
)

func main() {
	fmt.Println("=== Ti Configuration and Authentication Demo ===\n")

	// ====================
	// Configuration Example
	// ====================
	fmt.Println("--- Configuration Management ---")

	// Create a config manager
	cfgManager, err := config.NewManager("")
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	fmt.Printf("Config file: %s\n", cfgManager.ConfigPath())

	// Load config
	cfg, err := cfgManager.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Add some settings
	cfg.SetSetting("theme", "dark")
	cfg.SetSetting("language", "en")
	cfg.SetSetting("editor", "vim")

	if err := cfgManager.Save(cfg); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}
	fmt.Println("✓ Settings saved")

	// Read settings
	if theme, ok := cfg.GetSetting("theme"); ok {
		fmt.Printf("  Theme: %s\n", theme)
	}
	fmt.Println()

	// ====================
	// Authentication Example
	// ====================
	fmt.Println("--- Authentication Management (Multiple Providers) ---")

	// Create auth store manager
	authManager, err := authstore.NewManager("")
	if err != nil {
		log.Fatalf("Failed to create auth manager: %v", err)
	}
	fmt.Printf("Auth file: %s\n", authManager.AuthPath())

	// Load auth store
	store, err := authManager.Load()
	if err != nil {
		log.Fatalf("Failed to load auth store: %v", err)
	}

	// Add credentials for multiple providers
	expiresAt := time.Now().Add(1 * time.Hour).Unix()

	store.SetCredentials("google", "google-access-token", "google-refresh-token", expiresAt)
	store.SetCredentials("github", "github-access-token", "github-refresh-token", expiresAt)
	store.SetCredentials("gitlab", "gitlab-access-token", "gitlab-refresh-token", expiresAt)

	if err := authManager.Save(store); err != nil {
		log.Fatalf("Failed to save auth store: %v", err)
	}
	fmt.Println("✓ Credentials saved for multiple providers\n")

	// List all providers
	fmt.Println("Registered providers:")
	for _, provider := range store.ListProviders() {
		fmt.Printf("  - %s", provider)
		if store.IsExpired(provider) {
			fmt.Print(" (expired)")
		}
		fmt.Println()
	}
	fmt.Println()

	// Get credentials for specific provider
	fmt.Println("Google credentials:")
	if googleCred, err := store.GetCredentials("google"); err == nil {
		fmt.Printf("  Access Token: %s\n", googleCred.AccessToken)
		fmt.Printf("  Refresh Token: %s\n", googleCred.RefreshToken)
		fmt.Printf("  Provider: %s\n", googleCred.Provider)
		fmt.Printf("  Expires At: %d\n", googleCred.ExpiresAt)
		fmt.Printf("  Updated At: %s\n", googleCred.UpdatedAt.Format(time.RFC3339))
	}
	fmt.Println()

	// Check if provider has credentials
	fmt.Println("Provider status:")
	fmt.Printf("  Has Google? %v\n", store.HasCredentials("google"))
	fmt.Printf("  Has Twitter? %v\n", store.HasCredentials("twitter"))
	fmt.Println()

	// Update access token (refresh scenario)
	fmt.Println("--- Token Refresh Example ---")
	newExpiresAt := time.Now().Add(2 * time.Hour).Unix()
	if err := store.UpdateAccessToken("google", "new-google-access-token", newExpiresAt); err == nil {
		fmt.Println("✓ Google access token refreshed")
		if googleCred, err := store.GetCredentials("google"); err == nil {
			fmt.Printf("  New Access Token: %s\n", googleCred.AccessToken)
			fmt.Printf("  Refresh Token (unchanged): %s\n", googleCred.RefreshToken)
		}
	}
	fmt.Println()

	// Remove credentials for a provider
	fmt.Println("--- Remove Provider ---")
	store.RemoveCredentials("gitlab")
	authManager.Save(store)
	fmt.Println("✓ GitLab credentials removed")
	fmt.Printf("  Remaining providers: %v\n", store.ListProviders())
	fmt.Println()

	// Reload to demonstrate persistence
	fmt.Println("--- Persistence Check ---")
	reloadedStore, err := authManager.Load()
	if err != nil {
		log.Fatalf("Failed to reload auth store: %v", err)
	}
	fmt.Printf("  Providers after reload: %v\n", reloadedStore.ListProviders())
	fmt.Println()

	// ====================
	// Cleanup Examples
	// ====================
	fmt.Println("--- Cleanup Options ---")

	// Clear all auth
	fmt.Println("To clear all authentication:")
	fmt.Println("  store.ClearAll()")
	fmt.Println("  authManager.Save(store)")

	// Delete files
	fmt.Println("\nTo delete config/auth files:")
	fmt.Println("  cfgManager.Delete()")
	fmt.Println("  authManager.Delete()")

	// Uncomment to actually clean up:
	// store.ClearAll()
	// authManager.Save(store)
	// cfgManager.Delete()
	// authManager.Delete()

	fmt.Println("\n=== Demo Complete ===")
}
