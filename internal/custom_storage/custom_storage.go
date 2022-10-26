package custom_storage

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	"github.com/go-git/go-git/v5/storage/memory"
)

type ObjectStorage = filesystem.ObjectStorage
type IndexStorage = memory.IndexStorage
type ShallowStorage = CustomShallowStorage
type ConfigStorage = CustomConfigStorage
type ModuleStorage = CustomModuleStorage
type ReferenceStorage = CustomReferenceStorage

type CustomStorage struct {
	fs  billy.Filesystem
	dir *dotgit.DotGit

	ObjectStorage
	ReferenceStorage
	IndexStorage
	ShallowStorage
	ConfigStorage
	ModuleStorage
}

// NewCustomStorage returns a new CustomStorage that uses an in-memory Index storage
// backed by a given `fs.Filesystem` and cache.
func NewCustomStorage(fs billy.Filesystem, cache cache.Object) (storage *CustomStorage) {
	repoDir := dotgit.NewWithOptions(fs, dotgit.Options{})

	return &CustomStorage{
		fs:  fs,
		dir: repoDir,

		ObjectStorage:    *filesystem.NewObjectStorageWithOptions(repoDir, cache, filesystem.Options{}),
		ReferenceStorage: CustomReferenceStorage{dir: repoDir},
		IndexStorage:     memory.IndexStorage{},
		ShallowStorage:   CustomShallowStorage{dir: repoDir},
		ConfigStorage:    CustomConfigStorage{dir: repoDir},
		ModuleStorage:    CustomModuleStorage{dir: repoDir},
	}
}

// Filesystem returns the underlying filesystem
func (s *CustomStorage) Filesystem() billy.Filesystem {
	return s.fs
}
