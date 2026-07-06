package planner

import (
	"regexp"
	"strconv"

	"github.com/JavierMunioz/IAMXFREE/internal/inspection"
	"github.com/JavierMunioz/IAMXFREE/internal/models"
)

// nodeFrontendFrameworks / nodeBackendFrameworks are the frameworks this
// planner is confident enough to map to a models.ApplicationType. Hybrid
// meta-frameworks (Next.js, Nuxt, Astro can all serve as either frontend or
// backend/monolith) are deliberately left unclassified rather than guessed.
var nodeFrontendFrameworks = map[models.Framework]bool{
	models.FrameworkReact:   true,
	models.FrameworkVue:     true,
	models.FrameworkAngular: true,
	models.FrameworkSvelte:  true,
}

var nodeBackendFrameworks = map[models.Framework]bool{
	models.FrameworkExpress: true,
	models.FrameworkNestJS:  true,
}

var startScriptPortRe = regexp.MustCompile(`(?:--port[= ]|-p\s+|PORT=)(\d{2,5})`)

type nodePlanner struct{}

// NewNodePlanner returns a Planner for Node.js projects.
func NewNodePlanner() Planner {
	return &nodePlanner{}
}

func (p *nodePlanner) Name() string { return "node" }

func (p *nodePlanner) CanPlan(detection inspection.Detection) bool {
	return detection.Type == inspection.ProjectTypeNode
}

func (p *nodePlanner) Plan(detection inspection.Detection, result inspection.Result) DeploymentPlan {
	plan := DeploymentPlan{
		ProjectType:    detection.Type,
		SuggestedName:  detection.Name,
		Framework:      detection.Framework,
		Runtime:        detection.Runtime,
		PackageManager: detection.PackageManager,
		MatchedFiles:   append([]string(nil), detection.MatchedFiles...),
		Dependencies:   append([]string(nil), detection.Dependencies...),
		// InstallCommand/BuildCommand reuse the Node detector's own
		// suggestions: it already knows never to invent a script that
		// isn't in package.json, so this planner does not re-derive them.
		InstallCommand: detection.Suggested.Install,
		BuildCommand:   detection.Suggested.Build,
	}

	if plan.SuggestedName == "" {
		plan.Warnings = append(plan.Warnings, `application name could not be determined; package.json has no "name" field`)
	}

	if plan.PackageManager == "" {
		plan.Warnings = append(plan.Warnings, "package manager could not be determined; no lockfile was found")
	}

	if _, hasBuildScript := detection.Scripts["build"]; !hasBuildScript {
		plan.Notes = append(plan.Notes, `no "build" script found in package.json; a build step is not always required`)
	}

	if startScript, hasStartScript := detection.Scripts["start"]; hasStartScript {
		plan.StartCommand = detection.Suggested.Start
		if port, ok := inferPortFromScript(startScript); ok {
			plan.SuggestedPort = port
		} else {
			plan.Warnings = append(plan.Warnings, `port could not be inferred from the "start" script; specify one manually`)
		}
	} else {
		plan.Warnings = append(plan.Warnings, `no "start" script found in package.json; a start command could not be suggested`)
	}

	plan.Type = classifyNodeApplicationType(detection.Framework)
	if plan.Type == "" {
		plan.Notes = append(plan.Notes, "application type could not be confidently classified from the detected framework")
	}

	for _, sibling := range result.Detections {
		if sibling.Type == inspection.ProjectTypeDocker {
			plan.Notes = append(plan.Notes, "a Dockerfile or Compose file was also found alongside this Node project")
			break
		}
	}

	plan.Confidence = downgradeIfWarned(detection.Confidence, plan.Warnings)

	return plan
}

func classifyNodeApplicationType(framework models.Framework) models.ApplicationType {
	switch {
	case nodeFrontendFrameworks[framework]:
		return models.ApplicationTypeFrontend
	case nodeBackendFrameworks[framework]:
		return models.ApplicationTypeBackend
	default:
		return ""
	}
}

func inferPortFromScript(script string) (int, bool) {
	m := startScriptPortRe.FindStringSubmatch(script)
	if m == nil {
		return 0, false
	}
	port, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return port, true
}

// downgradeIfWarned lowers base by one step when the plan ended up with
// warnings — those reflect gaps in the overall proposal, not just in the
// underlying detection's file-parsing confidence.
func downgradeIfWarned(base inspection.Confidence, warnings []string) inspection.Confidence {
	if len(warnings) == 0 {
		return base
	}
	switch base {
	case inspection.ConfidenceHigh:
		return inspection.ConfidenceMedium
	default:
		return inspection.ConfidenceLow
	}
}
