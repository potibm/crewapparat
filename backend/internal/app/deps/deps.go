// Package deps hält Abhängigkeiten fest, die das Grundgerüst vorlädt, aber
// noch kein Produktivcode benutzt.
//
// Grund: das Projekt muss auch ohne Netzwerk baubar sein (Zugfahrt, CI-Cache
// ohne Warm-up). `go mod tidy` berücksichtigt alle Build-Tags, trägt diese
// Imports aber nur, solange die Datei existiert. Sobald die Fachlogik die
// Pakete direkt importiert, kann dieses Paket hier gelöscht werden.
//
// Bauen mit: go build -tags deps ./...
package deps

import (
	_ "github.com/aws/aws-sdk-go-v2/service/sesv2"
	_ "github.com/google/uuid"
)
