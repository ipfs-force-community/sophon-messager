package version

var (
	CurrentCommit string

	BuildVersion = "1.12.1-rc1"

	Version = BuildVersion + CurrentCommit
)
