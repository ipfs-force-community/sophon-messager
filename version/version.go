package version

var (
	CurrentCommit string

	BuildVersion = "1.17.0"

	Version = BuildVersion + CurrentCommit
)
