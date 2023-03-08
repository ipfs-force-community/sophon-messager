package version

var (
	CurrentCommit string

	BuildVersion = "1.10.1"

	Version = BuildVersion + CurrentCommit
)
