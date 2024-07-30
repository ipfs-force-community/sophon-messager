package version

var (
	CurrentCommit string

	BuildVersion = "1.16.0"

	Version = BuildVersion + CurrentCommit
)
