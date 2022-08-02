package version

var (
	CurrentCommit string

	BuildVersion = "v1.6.0"

	Version = BuildVersion + CurrentCommit
)
