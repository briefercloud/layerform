package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ergomake/layerform/internal/terraform"
)

func TestState_StateTerraform(t *testing.T) {
	t.Run("returns the *teraform.State from inside", func(t *testing.T) {
		tfState := &terraform.State{
			Resources: []terraform.StateResource{{
				Module: "mod",
				Type:   "type",
				Name:   "name",
			}},
			Bytes: []byte{1, 2, 3},
		}
		s := &State{tfState}

		assert.Equal(t, tfState, s.Terraform())
	})

	t.Run("returns nil when State itself is nil", func(t *testing.T) {
		var s *State

		assert.Nil(t, s.Terraform())
	})
}
