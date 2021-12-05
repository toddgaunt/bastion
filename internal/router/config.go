package router

// Config contains all configuration for a Monastery website's commonindex and
// content pages, such as website title, css style, and which articles are
// pinned to the navigation bar instead of being indexed.
type Config struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Style        string            `json:"style"`
	Pinned       map[string]string `json:"pinned"`
	ScanInterval int
}
