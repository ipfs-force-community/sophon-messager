package version

var (
	CurrentCommit string

	BuildVersion = "1.12.0"

	Version = BuildVersion + CurrentCommit
)
