package paramsbuilder

import (
	"errors"

	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

var ErrMissingWorkspace = errors.New("missing workspace name")

type WorkspaceHolder interface {
	GiveWorkspace() *Workspace
}

// Workspace params sets up varying workspace name.
type Workspace struct {
	Name string
}

func (p *Workspace) GiveWorkspace() *Workspace {
	return p
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
func (p *Workspace) GetSubstitutionPlan() catalogreplacer.SubstitutionPlan {
	return catalogreplacer.SubstitutionPlan{
		From: catalogreplacer.VariableWorkspace,
		To:   p.Name,
	}
}
