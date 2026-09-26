package version

var (
	CurrentCommit string

	BuildVersion = "1.21.0-rc1"

	Version = BuildVersion + CurrentCommit
)
