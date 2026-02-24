package project

// Run all non-integration tests:
//   go generate -run testunit ./...
//
// Run MailHog integration tests only:
//   go generate -run testintegration ./...
//
//go:generate -command testunit go test
//go:generate -command testintegration go test -tags=integration
//go:generate testunit ./...
//go:generate testintegration ./internal/service/sender/mailsender
