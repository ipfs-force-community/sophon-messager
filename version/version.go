package version

var (
	CurrentCommit string

	BuildVersion = "1.17.1"

	Version = BuildVersion + CurrentCommit
)
