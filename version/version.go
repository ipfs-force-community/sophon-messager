package version

var (
	CurrentCommit string

	BuildVersion = "1.15.0"

	Version = BuildVersion + CurrentCommit
)
