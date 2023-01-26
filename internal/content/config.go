package content

// Config contains all configuration for a bastion server's router.
type Config struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Style        string `json:"style"`
	ScanInterval int    `json:"scan_interval"`
}
