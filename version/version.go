package version

var (
	CurrentCommit string

	BuildVersion = "1.14.0"

	Version = BuildVersion + CurrentCommit
)
