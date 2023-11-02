package version

var (
	CurrentCommit string

	BuildVersion = "1.14.0-rc2"

	Version = BuildVersion + CurrentCommit
)
