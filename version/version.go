package version

var (
	CurrentCommit string

	BuildVersion = "1.19.0"

	Version = BuildVersion + CurrentCommit
)
