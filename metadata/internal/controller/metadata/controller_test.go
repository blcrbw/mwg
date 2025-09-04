package metadata

import (
	"context"
	"errors"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
	"testing"

	gen "mmoviecom/gen/mock/metadata/repository"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestController(t *testing.T) {
	tests := []struct {
		name         string
		expRepoRes   *model.Metadata
		expRepoErr   error
		cacheRepoRes *model.Metadata
		cacheRepoErr error
		cachePutCall bool
		wantRes      *model.Metadata
		wantErr      error
	}{
		{
			name:         "not found",
			expRepoErr:   repository.ErrNotFound,
			cacheRepoErr: repository.ErrNotFound,
			cachePutCall: false,
			wantErr:      ErrNotFound,
		},
		{
			name:         "unexpected error",
			expRepoErr:   errors.New("unexpected error"),
			cacheRepoErr: repository.ErrNotFound,
			cachePutCall: false,
			wantErr:      errors.New("unexpected error"),
		},
		{
			name:         "success",
			expRepoRes:   &model.Metadata{},
			cacheRepoErr: repository.ErrNotFound,
			cachePutCall: true,
			wantRes:      &model.Metadata{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repoMock := gen.NewMockmetadataRepository(ctrl)
			cacheMock := gen.NewMockmetadataRepository(ctrl)
			c := New(repoMock, cacheMock)
			ctx := context.Background()
			id := "id"
			repoMock.EXPECT().Get(ctx, id).Return(tt.expRepoRes, tt.expRepoErr)
			cacheMock.EXPECT().Get(ctx, id).Return(tt.cacheRepoRes, tt.cacheRepoErr)
			if tt.cachePutCall {
				cacheMock.EXPECT().Put(ctx, id, tt.expRepoRes).Return(nil)
			}
			res, err := c.Get(ctx, id)
			assert.Equal(t, tt.wantRes, res, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}
