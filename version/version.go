package version

var (
	CurrentCommit string

	BuildVersion = "1.9.0-rc1"

	Version = BuildVersion + CurrentCommit
)
