package file

import (
	"github.com/isd-sgcu/rnkm65-file/src/app/model/file"
	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
}

func (r *RepositoryMock) FindByOwnerID(id string, in *file.File) error {
	args := r.Called(id, in)

	if args.Get(0) != nil {
		*in = *args.Get(0).(*file.File)
	}

	return args.Error(1)
}

func (r *RepositoryMock) CreateOrUpdate(in *file.File) error {
	args := r.Called(in.OwnerID)

	if args.Get(0) != nil {
		*in = *args.Get(0).(*file.File)
	}

	return args.Error(1)
}

func (r *RepositoryMock) Delete(id string) error {
	args := r.Called(id)

	return args.Error(0)
}
