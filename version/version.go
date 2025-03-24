package version

var (
	CurrentCommit string

	BuildVersion = "1.18.0-rc2"

	Version = BuildVersion + CurrentCommit
)
