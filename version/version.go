package version

var (
	CurrentCommit string

	BuildVersion = "1.9.1"

	Version = BuildVersion + CurrentCommit
)
