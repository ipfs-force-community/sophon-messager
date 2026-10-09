package version

var (
	CurrentCommit string

	BuildVersion = "1.21.0"

	Version = BuildVersion + CurrentCommit
)
