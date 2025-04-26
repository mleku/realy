package codegen

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestGen(t *testing.T) {

	parent := path.Join(toolbox.CallerDirectory(3), "test_data")
	os.Setenv("AST_DEBUG","0")
	var useCases = []struct {
		description string
		options     *Options
		hasError    bool
	}{

		{
			description: "map  types",
			options: &Options{
				Source: path.Join(parent, "maps"),
				Types:  []string{"Message"},
				Dest:   path.Join(parent, "maps", "message_enc.go"),
			},
		},
		{
			description: "slice  types",
			options: &Options{
				Source: path.Join(parent, "slices"),
				Types:  []string{"Message"},
				Dest:   path.Join(parent, "slices", "message_enc.go"),
			},
		},
		{
			description: "message  types",
			options: &Options{
				Source: path.Join(parent, "messages"),
				Types:  []string{"Message"},
				Dest:   path.Join(parent, "messages", "message_enc.go"),
			},
		},
		{
			description: "basic struct code generation",
			options: &Options{
				Source: path.Join(parent, "basic_struct"),
				Types:  []string{"Message"},
				Dest:   path.Join(parent, "basic_struct", "basic_struct.go"),
			},
		},
		{
			description: "basic aliased types",
			options: &Options{
				Source: path.Join(parent, "primitive_alias"),
				Types:  []string{"Message"},
				Dest:   path.Join(parent, "primitive_alias", "primitive_alias.go"),
			},
		},
	}

	for _, useCase := range useCases[0:1] {
		err := Generate(useCase.options)
		assert.Nil(t, err, useCase.hasError)

	}
}
