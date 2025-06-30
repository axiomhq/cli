package mcptools

import "context"

// Tool represents a tool that can be used to perform a specific task.
type Tool[TIn, TOut any] struct {
	// Name is the name of the tool, for programmatic identification.
	Name string
	// Title is the title of the tool, for display purposes.
	Title string
	// Description is the description of the tool
	Description string
	// Handler is the function that will be called when the tool is executed.
	Handler func(context.Context, TIn) (TOut, error)
}
