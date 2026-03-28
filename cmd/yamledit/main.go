package main

import (
	"github.com/suzuki-shunsuke/yamledit/pkg/cli"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var version = ""

func main() {
	urfave.Main("yamledit", version, cli.Run)
}
