package version

// Version and Channel are injected into release builds using -ldflags.
var (
	Version = "v0.0.0"
	Channel = "development"
)
