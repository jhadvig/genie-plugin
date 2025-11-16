package version

var version = "0.0.1" // -ldflags "-X github.com/inecas/obs-mcp/version.version=$(VERSION)"

// GetVersion returns the current version of the application
func GetVersion() string {
	return version
}
