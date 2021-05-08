package burner

var (
	version   = "v0.0.0-SNAPSHOT"
	buildDate = "0000-00-00 00:00:00.00Z"
	builtBy   = "-"

	versionInfo = VersionInfo{
		Version:   version,
		BuildDate: buildDate,
		BuiltBy:   builtBy,
	}
)

type VersionInfo struct {
	Version   string
	BuildDate string
	BuiltBy   string
}

func GetVersionInfo() VersionInfo {
	return versionInfo
}
