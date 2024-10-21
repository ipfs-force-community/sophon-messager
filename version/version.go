package version

var (
	CurrentCommit string

	BuildVersion = "1.17.0-rc1"

	Version = BuildVersion + CurrentCommit
)
