package models

import "github.com/google/go-github/github"

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (f *File) GetTreeEntry() *github.TreeEntry {
	return &github.TreeEntry{
		Path:    github.String(f.Path),
		Content: github.String(f.Content),
		Mode:    github.String("100644"),
	}
}
