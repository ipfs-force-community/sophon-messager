package version

var (
	CurrentCommit string

	BuildVersion = "v1.8.0-rc3"

	Version = BuildVersion + CurrentCommit
)
