package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidRepository is returned when a repository reference is not "owner/name".
var ErrInvalidRepository = errors.New(`repository must have the form "owner/name"`)

// Repository identifies a code repository by owner and name.
type Repository struct {
	Owner string
	Name  string
}

// ParseRepository reads an "owner/name" reference.
func ParseRepository(ref string) (Repository, error) {
	owner, name, ok := strings.Cut(strings.TrimSpace(ref), "/")
	if !ok || owner == "" || name == "" || strings.Contains(name, "/") {
		return Repository{}, fmt.Errorf("%w, got %q", ErrInvalidRepository, ref)
	}
	return Repository{Owner: owner, Name: name}, nil
}

func (r Repository) String() string {
	return r.Owner + "/" + r.Name
}

// ChangedFile is one file touched by a pull request.
type ChangedFile struct {
	Path      string
	Additions int
	Deletions int
}

// Lines returns the number of modified lines in the file.
func (f ChangedFile) Lines() int {
	return f.Additions + f.Deletions
}

// PullRequest is a proposed change and the files it touches.
type PullRequest struct {
	Repository Repository
	Number     int
	Title      string
	Author     string
	URL        string
	Files      []ChangedFile
}
