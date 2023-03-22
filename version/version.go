package version

var (
	CurrentCommit string

	BuildVersion = "1.10.2"

	Version = BuildVersion + CurrentCommit
)
