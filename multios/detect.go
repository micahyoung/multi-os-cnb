package multios

import (
	"github.com/paketo-buildpacks/packit"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{
						Name: context.BuildpackInfo.Name,
					},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: context.BuildpackInfo.Name,
					},
				},
			},
		}, nil
	}
}
