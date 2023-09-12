package version

var (
	CurrentCommit string

	BuildVersion = "1.13.0"

	Version = BuildVersion + CurrentCommit
)
