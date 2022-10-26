package custom_storage

import (
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

type CustomModuleStorage struct {
	dir *dotgit.DotGit
}

func (s *CustomModuleStorage) Module(name string) (storage.Storer, error) {
	fs, err := s.dir.Module(name)
	if err != nil {
		return nil, err
	}

	return filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), nil
}
