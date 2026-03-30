package core

// Document is the plain Ferret-facing robots.txt representation.
type Document struct {
	Host     *string  `ferret:"host" json:"host"`
	Groups   []Group  `ferret:"groups" json:"groups"`
	Sitemaps []string `ferret:"sitemaps" json:"sitemaps"`
}

// Group represents a single robots user-agent group.
type Group struct {
	CrawlDelay *float64 `ferret:"crawlDelay" json:"crawlDelay"`
	UserAgents []string `ferret:"userAgents" json:"userAgents"`
	Allow      []string `ferret:"allow" json:"allow"`
	Disallow   []string `ferret:"disallow" json:"disallow"`
}

// MatchResult reports the effective rule chosen for a path and user-agent.
type MatchResult struct {
	Directive *string `ferret:"directive" json:"directive"`
	Pattern   *string `ferret:"pattern" json:"pattern"`
	UserAgent string  `ferret:"userAgent" json:"userAgent"`
	Allowed   bool    `ferret:"allowed" json:"allowed"`
}
