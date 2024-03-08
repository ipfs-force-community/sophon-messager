package version

var (
	CurrentCommit string

	BuildVersion = "1.15.0-rc1"

	Version = BuildVersion + CurrentCommit
)
