package version

var (
	CurrentCommit string

	BuildVersion = "v1.8.0-rc1"

	Version = BuildVersion + CurrentCommit
)
