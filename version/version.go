package version

var (
	CurrentCommit string

	BuildVersion = "1.14.0-rc3"

	Version = BuildVersion + CurrentCommit
)
