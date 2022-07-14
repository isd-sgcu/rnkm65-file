package cache

import (
	dto "github.com/isd-sgcu/rnkm65-file/src/app/dto/file"
	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
	V map[string]interface{}
}

func (t *RepositoryMock) SaveCache(key string, v interface{}, ttl int) error {
	args := t.Called(key, v.(*dto.CacheFile).Url, ttl)

	t.V[key] = v

	return args.Error(0)
}

func (t *RepositoryMock) GetCache(key string, v interface{}) error {
	args := t.Called(key, v)

	if args.Get(0) != nil {
		*v.(*dto.CacheFile) = *args.Get(0).(*dto.CacheFile)
	}

	return args.Error(1)
}
