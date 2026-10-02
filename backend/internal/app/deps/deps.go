// Package deps pins dependencies that the platform pre-fetches but no
// product code uses yet.
//
// Reason: the project has to build without network access (train ride, cold CI
// cache). `go mod tidy` considers all build tags, so it keeps these imports
// for as long as this file exists. Once the domain logic imports the packages
// directly, this package can be deleted.
//
// Build with: go build -tags deps ./...
package deps

import (
	// AWS SES v2 for the magic-link invitations (CONCEPT.md 4.1).
	_ "github.com/aws/aws-sdk-go-v2/service/sesv2"
)
