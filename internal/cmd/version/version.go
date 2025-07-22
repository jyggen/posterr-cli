package version

import (
	"fmt"
	"github.com/jyggen/posterr-cli/internal"
)

type Command struct{}

func (cmd *Command) Run() error {
	fmt.Println(internal.Version())
	return nil
}
