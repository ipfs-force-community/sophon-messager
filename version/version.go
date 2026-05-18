package version

var (
	CurrentCommit string

	BuildVersion = "1.20.0"

	Version = BuildVersion + CurrentCommit
)
