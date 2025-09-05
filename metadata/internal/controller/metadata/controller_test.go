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

func TestControllerGet(t *testing.T) {
	tests := []struct {
		name         string
		expRepoRes   *model.Metadata
		expRepoErr   error
		cacheRepoRes *model.Metadata
		cacheRepoErr error
		cachePutErr  error
		cachePutCall bool
		repGetCall   bool
		cacheGetCall bool
		wantRes      *model.Metadata
		wantErr      error
	}{
		{
			name:         "not found",
			expRepoErr:   repository.ErrNotFound,
			cacheRepoErr: repository.ErrNotFound,
			cachePutCall: false,
			repGetCall:   true,
			cacheGetCall: true,
			wantErr:      ErrNotFound,
		},
		{
			name:         "unexpected error",
			expRepoErr:   errors.New("unexpected error"),
			cacheRepoErr: repository.ErrNotFound,
			cachePutCall: false,
			repGetCall:   true,
			cacheGetCall: true,
			wantErr:      errors.New("unexpected error"),
		},
		{
			name:         "success",
			expRepoRes:   &model.Metadata{},
			cacheRepoErr: repository.ErrNotFound,
			cachePutCall: true,
			repGetCall:   true,
			cacheGetCall: true,
			wantRes:      &model.Metadata{},
		},
		{
			name:         "found in cache",
			expRepoRes:   &model.Metadata{},
			cacheRepoRes: &model.Metadata{},
			cachePutCall: false,
			repGetCall:   false,
			cacheGetCall: true,
			wantRes:      &model.Metadata{},
		},
		{
			name:         "cache put error",
			expRepoRes:   &model.Metadata{},
			cacheRepoErr: repository.ErrNotFound,
			cachePutErr:  errors.New("unexpected error"),
			cachePutCall: true,
			repGetCall:   true,
			cacheGetCall: true,
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
			if tt.repGetCall {
				repoMock.EXPECT().Get(ctx, id).Return(tt.expRepoRes, tt.expRepoErr)
			}
			if tt.cacheGetCall {
				cacheMock.EXPECT().Get(ctx, id).Return(tt.cacheRepoRes, tt.cacheRepoErr)
			}
			if tt.cachePutCall {
				cacheMock.EXPECT().Put(ctx, id, tt.expRepoRes).Return(tt.cachePutErr)
			}
			res, err := c.Get(ctx, id)
			assert.Equal(t, tt.wantRes, res, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}

func TestControllerPut(t *testing.T) {
	tests := []struct {
		name       string
		expRepoErr error
		wantErr    error
	}{
		{
			name:       "unexpected error",
			expRepoErr: errors.New("unexpected error"),
			wantErr:    errors.New("unexpected error"),
		},
		{
			name: "success",
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
			m := model.Metadata{
				ID:          "id",
				Title:       "title",
				Description: "description",
				Director:    "director",
			}
			repoMock.EXPECT().Put(ctx, m.ID, &m).Return(tt.expRepoErr)
			err := c.Put(ctx, m.ID, &m)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}
