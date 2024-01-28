package site

// Provider is an interface for sites provider, it's main purpose is to look up sites by their names
type Provider interface {
	GetSite(name string) (Site, bool)
	GetAllSiteNames() []string
	GetNewSiteNamesSinceInit() []string
}
