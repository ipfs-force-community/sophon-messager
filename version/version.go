package version

var (
	CurrentCommit string

	BuildVersion = "1.19.0-rc1"

	Version = BuildVersion + CurrentCommit
)
