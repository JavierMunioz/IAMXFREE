package models

// Runtime identifies the language/runtime an application executes on. Like
// Framework, this is an open string type so runtimes outside the known set
// can still be stored.
type Runtime string

const (
	RuntimeNode   Runtime = "node"
	RuntimeBun    Runtime = "bun"
	RuntimeDeno   Runtime = "deno"
	RuntimePython Runtime = "python"
	RuntimePHP    Runtime = "php"
	RuntimeGo     Runtime = "go"
	RuntimeJava   Runtime = "java"
	RuntimeRust   Runtime = "rust"
	RuntimeOther  Runtime = "other"
)
