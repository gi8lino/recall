package runner

// Config contains resolved runtime configuration shared by commands.
type Config struct {
	DataDir      string   // DataDir is the directory where recall stores its own data.
	HistoryFiles []string // HistoryFiles contains optional custom history files to scan.
	HistoryLimit int      // HistoryLimit is the maximum number of recent history entries to read, with zero meaning unlimited.
	Limit        int      // Limit is the maximum number of search results to print, with zero meaning unlimited.
}
