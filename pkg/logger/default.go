package logger

import (
	"fmt"
	"go.uber.org/zap"
	"os"
)

var (
	Default *zap.Logger
)

func init() {
	var err error
	Default, err = zap.NewProduction()
	if err != nil {
		fmt.Fprint(os.Stderr, "error loading logger: ", err)
		os.Exit(1)
	}
}
