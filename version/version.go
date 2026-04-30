package version

var (
	CurrentCommit string

	BuildVersion = "1.20.0-rc1"

	Version = BuildVersion + CurrentCommit
)
