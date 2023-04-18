package version

var (
	CurrentCommit string

	BuildVersion = "1.11.0-rc1"

	Version = BuildVersion + CurrentCommit
)
