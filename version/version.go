package version

var (
	CurrentCommit string

	BuildVersion = "1.10.0-rc3"

	Version = BuildVersion + CurrentCommit
)
