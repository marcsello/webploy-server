package adapters

import (
	"fmt"
	"go.uber.org/zap"
	"log"
)

type LogAdapter struct {
	L *zap.Logger
}

func (la LogAdapter) Printf(format string, a ...interface{}) {
	if la.L != nil {
		msg := fmt.Sprintf(format, a...)
		la.L.Info(msg)
	} else {
		log.Printf(format, a...)
	}
}
