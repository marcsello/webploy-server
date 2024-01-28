package info

type InfoTransaction func(i *DeploymentInfo) error

type InfoProvider interface {
	Tx(readonly bool, txFunc InfoTransaction) error
}
