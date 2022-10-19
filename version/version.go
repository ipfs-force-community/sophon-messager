package version

var (
	CurrentCommit string

	BuildVersion = "v1.8.0-rc2"

	Version = BuildVersion + CurrentCommit
)
