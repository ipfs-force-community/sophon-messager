package version

var (
	CurrentCommit string

	BuildVersion = "1.16.0-rc1"

	Version = BuildVersion + CurrentCommit
)
