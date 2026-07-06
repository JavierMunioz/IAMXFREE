package models

// Framework identifies the framework an application is built with. The
// constants below cover the stacks IAMXFREE recognizes out of the box, but
// the type itself is an open string so unrecognized or custom frameworks can
// still be recorded, using FrameworkOther or any raw value.
type Framework string

const (
	FrameworkReact   Framework = "react"
	FrameworkVue     Framework = "vue"
	FrameworkAngular Framework = "angular"
	FrameworkNextJS  Framework = "nextjs"
	FrameworkNuxt    Framework = "nuxt"
	FrameworkAstro   Framework = "astro"
	FrameworkSvelte  Framework = "svelte"
	FrameworkExpress Framework = "express"
	FrameworkNestJS  Framework = "nestjs"
	FrameworkDjango  Framework = "django"
	FrameworkFlask   Framework = "flask"
	FrameworkFastAPI Framework = "fastapi"
	FrameworkLaravel Framework = "laravel"
	FrameworkSpring  Framework = "spring"
	FrameworkGo      Framework = "go"
	FrameworkRust    Framework = "rust"
	FrameworkDotNet  Framework = "dotnet"
	FrameworkOther   Framework = "other"
)
