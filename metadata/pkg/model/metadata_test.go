package model

import (
	gen "mmoviecom/gen"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestProtobufConversion(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Metadata conversion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			model := Metadata{
				ID:          "id",
				Title:       "title",
				Description: "description",
				Director:    "director",
			}
			genModel := gen.Metadata{
				Id:          "id",
				Title:       "title",
				Description: "description",
				Director:    "director",
			}

			m2p := MetadataToProto(&model)
			p2m := MetadataFromProto(&genModel)
			m2pDiff := cmp.Diff(m2p, &genModel, cmpopts.IgnoreUnexported(gen.Metadata{}, Metadata{}))
			assert.Equal(t, "", m2pDiff, tt.name)
			p2mDiff := cmp.Diff(p2m, &model, cmpopts.IgnoreUnexported(gen.Metadata{}, Metadata{}))
			assert.Equal(t, "", p2mDiff, tt.name)
		})
	}
}
