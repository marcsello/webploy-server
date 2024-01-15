package deployment

type StateProvider interface {
	IsFinished() bool
	Creator() string
}

type StateProviderLocalFile struct {
}
