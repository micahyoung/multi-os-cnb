package multios

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"os"
	"path/filepath"
)

//DependencyService for packit
type DependencyService interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//Build for packit
func Build(logger scribe.Logger, service DependencyService) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		goRootLayer, err := context.Layers.Get("go-root", packit.BuildLayer, packit.CacheLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Providing GO")
		goCompilerPath := filepath.Join(goRootLayer.Path, "go", "bin")

		if _, err := os.Stat(goCompilerPath); err != nil && os.IsNotExist(err) {
			dep, err := service.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), "go", "", context.Stack)
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Subprocess("Downloading GO")
			err = service.Install(dep, context.CNBPath, goRootLayer.Path)
			if err != nil {
				return packit.BuildResult{}, err
			}
		} else {
			logger.Subprocess("Reusing cached GO")
		}

		launchLayer, err := context.Layers.Get("go-launch", packit.LaunchLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// use custom tmp/cache path since TMP is not set on Windows (for some TBD reason)
		goTmp := filepath.Join(goRootLayer.Path, "tmp")
		if err := os.MkdirAll(goTmp, 0777); err != nil {
			return packit.BuildResult{}, err
		}

		launchBinPath := filepath.Join(launchLayer.Path, "bin")
		if err := os.MkdirAll(launchBinPath, 0777); err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Building")
		if err := pexec.NewExecutable("go").Execute(
			pexec.Execution{
				Args: []string{"build", "-o", launchBinPath, "./cmd/app"}, //TODO, search for any main.go and get proper package name
				Env: []string{
					"PATH=" + goCompilerPath,
					"GOPATH=" + filepath.Join(goRootLayer.Path, "go"),
					"GOCACHE=" + goTmp,
					"GOTMPDIR=" + goTmp,
				},
				Stdout: os.Stdout,
				Stderr: os.Stderr,
				Dir:    context.WorkingDir,
			}); err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan: context.Plan,
			Layers: []packit.Layer{
				launchLayer,
				goRootLayer,
			},
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: "app",
						Direct:  true,
					},
				},
			},
		}, nil
	}
}
