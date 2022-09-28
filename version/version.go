package version

var (
	CurrentCommit string

	BuildVersion = "v1.7.1"

	Version = BuildVersion + CurrentCommit
)
