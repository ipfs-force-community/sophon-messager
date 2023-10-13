package version

var (
	CurrentCommit string

	BuildVersion = "1.14.0-rc1"

	Version = BuildVersion + CurrentCommit
)
