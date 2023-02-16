package version

var (
	CurrentCommit string

	BuildVersion = "1.10.0-rc1"

	Version = BuildVersion + CurrentCommit
)
