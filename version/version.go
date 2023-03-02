package version

var (
	CurrentCommit string

	BuildVersion = "1.10.0"

	Version = BuildVersion + CurrentCommit
)
