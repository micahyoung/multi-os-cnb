package multios

import (
	"errors"
	"fmt"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//DependencyService for packit
type DependencyService interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//Build for packit
func Build(logger scribe.Logger, service DependencyService, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Process("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		// get targets from BP_GO_TARGETS
		var targets []string
		if val, ok := os.LookupEnv("BP_GO_TARGETS"); ok {
			targets = filepath.SplitList(val)
		}

		// if no BP_GO_TARGETS, default search in cmd/main + cmd/**/main.go
		if len(targets) == 0 {
			mainMatches, _ := filepath.Glob(filepath.Join(context.WorkingDir, "cmd", "main.go"))
			mainMatchesNested, _ := filepath.Glob(filepath.Join(context.WorkingDir, "cmd", "**", "main.go"))

			for _, match := range append(mainMatches, mainMatchesNested...) {
				targets = append(targets, strings.ReplaceAll(filepath.Dir(match), context.WorkingDir, "."))
			}
		}

		if len(targets) == 0 {
			return packit.BuildResult{}, errors.New("no main.go files could be found")
		}

		goCacheLayer, err := context.Layers.Get("go", packit.BuildLayer, packit.CacheLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Resolving Go version")
		dep, err := service.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), "go", "", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Action("Selected Go version: %s", dep.Version)

		goCompilerPath := filepath.Join(goCacheLayer.Path, "go", "bin")
		if _, err := os.Stat(goCompilerPath); err != nil && os.IsNotExist(err) {
			logger.Action("Downloading GO")
			err = service.Install(dep, context.CNBPath, goCacheLayer.Path)
			if err != nil {
				return packit.BuildResult{}, err
			}
		} else {
			logger.Subprocess("Reusing cached layer %s", filepath.Dir(goCompilerPath))
		}

		targetsLayer, err := context.Layers.Get("go-targets", packit.LaunchLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// use custom tmp/cache path since TMP is not set on Windows (for some TBD reason)
		goTmp := filepath.Join(goCacheLayer.Path, "tmp")
		if err := os.MkdirAll(goTmp, 0777); err != nil {
			return packit.BuildResult{}, err
		}

		launchBinPath := filepath.Join(targetsLayer.Path, "bin")
		if err := os.MkdirAll(launchBinPath, 0777); err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Executing build process")
		for _, target := range targets {
			logger.Action("Executing build process")
			args := []string{"build", "-o", launchBinPath, target}
			logger.Detail("Running 'go %s'", strings.Join(args, " "))

			duration, err := clock.Measure(func() error {
				if err := pexec.NewExecutable("go").Execute(
					pexec.Execution{
						Args: args,
						Env: []string{
							"PATH=" + goCompilerPath,
							"GOPATH=" + filepath.Join(goCacheLayer.Path, "go"),
							"GOCACHE=" + goTmp,
							"GOTMPDIR=" + goTmp,
						},
						Stdout: os.Stdout,
						Stderr: os.Stderr,
						Dir:    context.WorkingDir,
					}); err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				logger.Subdetail("Failed after %s", duration.Round(time.Millisecond))

				return packit.BuildResult{}, fmt.Errorf("failed to execute 'go build': %w", err)
			}

			logger.Subdetail("Completed in %s", duration.Round(time.Millisecond))
		}

		logger.Subprocess("Assigning launch processes")
		launchType := "web"
		launchTarget := targets[0]
		_, launchPathBinName := filepath.Split(launchTarget)

		logger.Action("%s: %s", launchType, filepath.Join(launchBinPath, launchTarget))

		return packit.BuildResult{
			Plan: context.Plan,
			Layers: []packit.Layer{
				targetsLayer,
				goCacheLayer,
			},
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    launchType,
						Command: launchPathBinName,
						Direct:  true,
					},
				},
			},
		}, nil
	}
}
