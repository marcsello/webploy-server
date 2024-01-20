package site

type ProviderMock struct {
	GetSiteFunc                  func(name string) (Site, bool)
	GetNewSiteNamesSinceInitFunc func() []string
}

func (p *ProviderMock) GetSite(name string) (Site, bool) {
	if p.GetSiteFunc != nil {
		return p.GetSiteFunc(name)
	}
	return nil, false
}

func (p *ProviderMock) GetNewSiteNamesSinceInit() []string {
	if p.GetNewSiteNamesSinceInitFunc != nil {
		return p.GetNewSiteNamesSinceInitFunc()
	}
	return nil
}
