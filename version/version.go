package version

var (
	CurrentCommit string

	BuildVersion = "1.18.0"

	Version = BuildVersion + CurrentCommit
)
