package tools

// Compile-time interface checks — if any tool stops implementing Tool, the
// build will fail with a clear error instead of a runtime panic.
var (
	// filesystem
	_ Tool = (*ReadFile)(nil)
	_ Tool = (*ReadFileRange)(nil)
	_ Tool = (*WriteFile)(nil)
	_ Tool = (*ListFiles)(nil)
	_ Tool = (*SearchCode)(nil)
	// go toolchain
	_ Tool = (*GoCommand)(nil)
	// go AST
	_ Tool = (*GoDefinitions)(nil)
	_ Tool = (*GoReadDefinition)(nil)
	// skills
	_ Tool = (*GetSkill)(nil)
	// memory
	_ Tool = (*MemorySave)(nil)
	_ Tool = (*MemoryGet)(nil)
	_ Tool = (*MemoryList)(nil)
	// delegation
	_ Tool = (*RunAgent)(nil)
	_ Tool = (*CreateTicket)(nil)
	_ Tool = (*GetTicket)(nil)
)
