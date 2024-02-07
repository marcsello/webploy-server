package site

// Provider is an interface for sites provider, it's main purpose is to look up sites by their names
// The number of sites will never change while the software is running, so the provider does not implement any complex transactional access to sites
type Provider interface {
	GetSite(name string) (Site, bool)
	GetAllSiteNames() []string
	GetNewSiteNamesSinceInit() []string
}
