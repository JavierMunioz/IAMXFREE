package application

import (
	"fmt"

	"github.com/JavierMunioz/IAMXFREE/internal/models"
	"github.com/JavierMunioz/IAMXFREE/internal/services"
	"github.com/JavierMunioz/IAMXFREE/internal/tui/wizard"
	"github.com/JavierMunioz/IAMXFREE/internal/validation"
)

// Step keys, shared between Steps and DraftFromResult.
const (
	KeyPath           = "path"
	KeyAnalysis       = "analysis"
	KeyName           = "name"
	KeyType           = "type"
	KeyFramework      = "framework"
	KeyRuntime        = "runtime"
	KeyPackageManager = "package_manager"
	KeyInstallCommand = "install_command"
	KeyBuildCommand   = "build_command"
	KeyStartCommand   = "start_command"
	KeyPort           = "port"
	KeyDomain         = "domain"
	KeyRepoURL        = "repo_url"
)

var typeChoices = []wizard.Choice{
	{Label: "Frontend", Value: string(models.ApplicationTypeFrontend)},
	{Label: "Backend", Value: string(models.ApplicationTypeBackend)},
	{Label: "Monolith", Value: string(models.ApplicationTypeMonolith)},
	{Label: "Worker", Value: string(models.ApplicationTypeWorker)},
	{Label: "Service", Value: string(models.ApplicationTypeService)},
	{Label: "API", Value: string(models.ApplicationTypeAPI)},
	{Label: "Microservice", Value: string(models.ApplicationTypeMicroservice)},
}

var frameworkChoices = []wizard.Choice{
	{Label: "React", Value: string(models.FrameworkReact)},
	{Label: "Vue", Value: string(models.FrameworkVue)},
	{Label: "Angular", Value: string(models.FrameworkAngular)},
	{Label: "Next.js", Value: string(models.FrameworkNextJS)},
	{Label: "Nuxt", Value: string(models.FrameworkNuxt)},
	{Label: "Astro", Value: string(models.FrameworkAstro)},
	{Label: "Svelte", Value: string(models.FrameworkSvelte)},
	{Label: "Express", Value: string(models.FrameworkExpress)},
	{Label: "NestJS", Value: string(models.FrameworkNestJS)},
	{Label: "Django", Value: string(models.FrameworkDjango)},
	{Label: "Flask", Value: string(models.FrameworkFlask)},
	{Label: "FastAPI", Value: string(models.FrameworkFastAPI)},
	{Label: "Laravel", Value: string(models.FrameworkLaravel)},
	{Label: "Spring", Value: string(models.FrameworkSpring)},
	{Label: "Go", Value: string(models.FrameworkGo)},
	{Label: "Rust", Value: string(models.FrameworkRust)},
	{Label: ".NET", Value: string(models.FrameworkDotNet)},
}

var runtimeChoices = []wizard.Choice{
	{Label: "Node", Value: string(models.RuntimeNode)},
	{Label: "Bun", Value: string(models.RuntimeBun)},
	{Label: "Deno", Value: string(models.RuntimeDeno)},
	{Label: "Python", Value: string(models.RuntimePython)},
	{Label: "PHP", Value: string(models.RuntimePHP)},
	{Label: "Go", Value: string(models.RuntimeGo)},
	{Label: "Java", Value: string(models.RuntimeJava)},
	{Label: "Rust", Value: string(models.RuntimeRust)},
}

func valueOrDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

// Steps builds the ordered step sequence for registering a new application.
// The project path is confirmed first, so an AnalysisStep can inspect it
// and let a services.ApplicationSetupService propose a configuration;
// Name/Type/Framework/Runtime/PackageManager/Install/Build/Start all
// pre-fill themselves from that proposal (via WithPrefill) but stay fully
// editable — nothing is ever locked, and Port/Domain/Repository are never
// pre-filled since the proposal doesn't cover them confidently enough.
// Adding a field to this wizard means adding one more entry here — the
// engine (internal/tui/wizard) never changes.
func Steps(setup services.ApplicationSetupService) []wizard.StepDef {
	path := wizard.NewTextStep("Path", "Local project path:", "/srv/apps/my-api", validation.Path())
	analysis := NewAnalysisStep(setup, path.Value)

	name := wizard.NewTextStep("Name", "Application name:", "my-api", validation.Required()).
		WithPrefill(func() string { return analysis.Proposal().SuggestedName })

	appType := wizard.NewChoiceStep("Type", "Application type:", typeChoices, false).
		WithPrefill(func() string { return string(analysis.Proposal().Type) })

	framework := wizard.NewChoiceStep("Framework", "Framework:", frameworkChoices, true).
		WithPrefill(func() string { return string(analysis.Proposal().Framework) })

	runtime := wizard.NewChoiceStep("Runtime", "Runtime:", runtimeChoices, true).
		WithPrefill(func() string { return string(analysis.Proposal().Runtime) })

	packageManager := wizard.NewTextStep("Package Manager", "Package manager (optional):", "npm", nil).
		WithPrefill(func() string { return analysis.Proposal().PackageManager })

	installCommand := wizard.NewTextStep("Install Command", "Install command (optional):", "npm install", nil).
		WithPrefill(func() string { return analysis.Proposal().InstallCommand })

	buildCommand := wizard.NewTextStep("Build Command", "Build command (optional):", "npm run build", nil).
		WithPrefill(func() string { return analysis.Proposal().BuildCommand })

	startCommand := wizard.NewTextStep("Start Command", "Start command (optional):", "npm start", nil).
		WithPrefill(func() string { return analysis.Proposal().StartCommand })

	port := wizard.NewTextStep("Port", "Internal port:", "3000", validation.Port())
	domain := wizard.NewTextStep("Domain", "Domain (optional):", "example.com", validation.Optional(validation.Domain()))
	repoURL := wizard.NewTextStep("Repository", "Git repository URL (optional):", "https://github.com/user/repo.git",
		validation.Optional(validation.GitRepository()))

	summary := wizard.NewSummaryStep("Summary", func() string {
		return fmt.Sprintf(
			"Name:            %s\nType:            %s\nFramework:       %s\nRuntime:         %s\nPackage manager: %s\nPath:            %s\nPort:            %s\nDomain:          %s\nRepository:      %s\nInstall:         %s\nBuild:           %s\nStart:           %s",
			name.Value(), appType.Value(), framework.Value(), runtime.Value(), valueOrDash(packageManager.Value()),
			path.Value(), port.Value(), valueOrDash(domain.Value()), valueOrDash(repoURL.Value()),
			valueOrDash(installCommand.Value()), valueOrDash(buildCommand.Value()), valueOrDash(startCommand.Value()),
		)
	})

	return []wizard.StepDef{
		{Key: KeyPath, Step: path},
		{Key: KeyAnalysis, Step: analysis},
		{Key: KeyName, Step: name},
		{Key: KeyType, Step: appType},
		{Key: KeyFramework, Step: framework},
		{Key: KeyRuntime, Step: runtime},
		{Key: KeyPackageManager, Step: packageManager},
		{Key: KeyInstallCommand, Step: installCommand},
		{Key: KeyBuildCommand, Step: buildCommand},
		{Key: KeyStartCommand, Step: startCommand},
		{Key: KeyPort, Step: port},
		{Key: KeyDomain, Step: domain},
		{Key: KeyRepoURL, Step: repoURL},
		{Key: "confirm", Step: summary},
	}
}
