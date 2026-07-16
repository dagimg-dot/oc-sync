package cli

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func Version() string   { return version }
func Commit() string    { return commit }
func BuildDate() string { return buildDate }
