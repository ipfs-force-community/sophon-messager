package version

var (
	CurrentCommit string

	BuildVersion = "1.9.0"

	Version = BuildVersion + CurrentCommit
)
