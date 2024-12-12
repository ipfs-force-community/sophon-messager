package version

var (
	CurrentCommit string

	BuildVersion = "1.18.0-rc1"

	Version = BuildVersion + CurrentCommit
)
