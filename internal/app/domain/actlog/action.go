package actlog

import "slices"

type Action string

func (a Action) String() string    { return string(a) }
func (a Action) IsValid() bool     { return slices.Contains(actions, a) }
func (a Action) Options() []Action { return actions }

const (
	ActionLogin Action = "auth-login"

	ActionGrantAccess  Action = "create-access-key"
	ActionRevokeAccess Action = "delete-access-key"

	ActionClearanceCheck Action = "immigration-clearance-check"
)

var actions = []Action{
	ActionLogin,
	ActionGrantAccess,
	ActionRevokeAccess,
	ActionClearanceCheck,
}
