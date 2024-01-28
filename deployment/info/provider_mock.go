package info

type InfoProviderMock struct {
	TxFn func(readonly bool, txFunc InfoTransaction) error
}

func (i *InfoProviderMock) Tx(readOnly bool, txFunc InfoTransaction) (err error) {
	if i.TxFn != nil {
		return i.TxFn(readOnly, txFunc)
	} else {
		return nil
	}
}
