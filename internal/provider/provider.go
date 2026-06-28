package provider

import "context"

type Request struct {
	Provider     string
	WorkingDir   string
	SystemPrompt string
	PlanPrompt   string
	ReviewPrompt string
	Diff         string
	ChangedFiles []string
	Lang         string
}

type Response struct {
	Content string
}

type Provider interface {
	Run(ctx context.Context, request Request) (Response, error)
}
