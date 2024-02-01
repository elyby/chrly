package version

var (
	version = "undefined"
	commit  = "unknown"
)

func Version() string {
	return version
}

func Commit() string {
	return commit
}
