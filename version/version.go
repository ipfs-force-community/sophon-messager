package version

var (
	CurrentCommit string

	BuildVersion = "1.10.0-rc2"

	Version = BuildVersion + CurrentCommit
)
