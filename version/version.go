package version

var (
	CurrentCommit string

	BuildVersion = "v1.7.0"

	Version = BuildVersion + CurrentCommit
)
