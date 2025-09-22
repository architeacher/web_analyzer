package main

import (
	"fmt"
	"log"
	"time"

	"aidanwoods.dev/go-paseto/v2"
)

func main() {
	// Generate a new key pair
	privateKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := privateKey.Public()

	fmt.Printf("Private Key (hex): %s\n", privateKey.ExportHex())
	fmt.Printf("Public Key (hex): %s\n", publicKey.ExportHex())
	fmt.Println()

	// Create token using standard PASETO methods only
	token := paseto.NewToken()

	now := time.Now()
	expirationTime := now.Add(time.Hour * 24 * 365 * 38) // 38 years

	// Set standard claims using PASETO's built-in methods
	token.SetIssuer("web-analyzer-service")
	token.SetSubject("test-user")
	token.SetAudience("web-analyzer-api")
	token.SetJti("proper-paseto-v4-token")
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(expirationTime)

	// Add custom scopes claim
	token.Set("scopes", []string{"analyze", "read"})

	// Sign the token
	signedToken := token.V4Sign(privateKey, nil)

	// Parse the token to see the actual claims structure
	parser := paseto.NewParser()
	parsedToken, err := parser.ParseV4Public(publicKey, signedToken, nil)
	if err != nil {
		log.Fatal("Failed to parse generated token:", err)
	}

	fmt.Printf("Generated PASETO v4 Token:\n%s\n\n", signedToken)
	fmt.Printf("Parsed Claims JSON:\n%s\n\n", string(parsedToken.ClaimsJSON()))

	fmt.Println("To use this token:")
	fmt.Printf("1. Update the public key in internal/adapters/middleware/auth.go line 39 to: %s\n", publicKey.ExportHex())
	fmt.Printf("2. Use the token in your curl command:\n")
	fmt.Printf("   curl -X POST https://api.web-analyzer.dev/v1/analyze \\\n")
	fmt.Printf("     -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("     -H \"Authorization: Bearer %s\" \\\n", signedToken)
	fmt.Printf("     -d '{\"url\": \"https://github.com/login\"}'\n")
}
