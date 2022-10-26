package custom_storage

// Copied verbatim from go-git/storage/filesystem

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

type CustomReferenceStorage struct {
	dir *dotgit.DotGit
}

func (r *CustomReferenceStorage) SetReference(ref *plumbing.Reference) error {
	return r.dir.SetRef(ref, nil)
}

func (r *CustomReferenceStorage) CheckAndSetReference(ref, old *plumbing.Reference) error {
	return r.dir.SetRef(ref, old)
}

func (r *CustomReferenceStorage) Reference(n plumbing.ReferenceName) (*plumbing.Reference, error) {
	return r.dir.Ref(n)
}

func (r *CustomReferenceStorage) IterReferences() (storer.ReferenceIter, error) {
	refs, err := r.dir.Refs()
	if err != nil {
		return nil, err
	}

	return storer.NewReferenceSliceIter(refs), nil
}

func (r *CustomReferenceStorage) RemoveReference(n plumbing.ReferenceName) error {
	return r.dir.RemoveRef(n)
}

func (r *CustomReferenceStorage) CountLooseRefs() (int, error) {
	return r.dir.CountLooseRefs()
}

func (r *CustomReferenceStorage) PackRefs() error {
	return r.dir.PackRefs()
}
