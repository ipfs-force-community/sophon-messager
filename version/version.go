package version

var (
	CurrentCommit string

	BuildVersion = "1.8.0"

	Version = BuildVersion + CurrentCommit
)
