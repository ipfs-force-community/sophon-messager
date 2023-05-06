package version

var (
	CurrentCommit string

	BuildVersion = "1.11.0"

	Version = BuildVersion + CurrentCommit
)
