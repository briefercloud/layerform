package main

import (
	"time"

	"github.com/ergomake/layerform/cmd/cli"

	"github.com/carlmjohnson/versioninfo"
)

func main() {
	cli.SetVersionInfo(versioninfo.Version, versioninfo.Revision, versioninfo.LastCommit.Format(time.RFC3339))
	cli.Execute()
}
