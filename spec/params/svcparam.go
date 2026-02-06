package params

// SvcParam holds common service parameters (e.g. from env or config).
type SvcParam struct {
	MySQLDSN     string
	CricAPIKey   string
	CricAPIBase  string
	ServerPort   string
}
