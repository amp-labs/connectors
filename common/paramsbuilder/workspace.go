package paramsbuilder

import "errors"

var ErrMissingWorkspace = errors.New("missing workspace name")

// Workspace params sets up varying workspace name.
type Workspace struct {
	Name string
}

func (p *Workspace) ValidateParams() error {
	if len(p.Name) == 0 {
		return ErrMissingWorkspace
	}

	return nil
}

func (p *Workspace) WithWorkspace(workspaceRef string) {
	p.Name = workspaceRef
}

// GetSubstitutionPlan of the workspace describes how to insert its value into string templates.
// This makes Workspace parameter a catalog variable.
func (p *Workspace) GetSubstitutionPlan() SubstitutionPlan {
	return SubstitutionPlan{
		From: variableWorkspace,
		To:   p.Name,
	}
}
