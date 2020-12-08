package main

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"os"

	"github.com/micahyoung/multi-os-cnb/multios"
)

func main() {
	packit.Run(
		multios.Detect(),
		multios.Build(
			scribe.NewLogger(os.Stdout),
			postal.NewService(cargo.NewTransport()),
			chronos.DefaultClock,
		),
	)
}
