package main

import (
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/suzuki-shunsuke/yamledit/pkg/cli"
)

var version = ""

func main() {
	urfave.Main("yamledit", version, cli.Run)
}
