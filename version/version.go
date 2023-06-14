package version

var (
	CurrentCommit string

	BuildVersion = "1.12.0-rc1"

	Version = BuildVersion + CurrentCommit
)
