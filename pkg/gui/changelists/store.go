package changelists

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// storePath is the location of the changelists file within a worktree's git
// dir. We keep it under the git dir (not the work tree) so it's naturally
// per-worktree and never shows up as an untracked file.
func storePath(worktreeGitDirPath string) string {
	return filepath.Join(worktreeGitDirPath, "lazygit", "changelists.json")
}

// Load reads the changelists for a worktree. A missing file is not an error: it
// yields an empty set, which renders identically to lazygit without
// changelists.
func Load(worktreeGitDirPath string) (*Set, error) {
	data, err := os.ReadFile(storePath(worktreeGitDirPath))
	if err != nil {
		if os.IsNotExist(err) {
			return &Set{}, nil
		}
		return nil, err
	}

	set := &Set{}
	if err := json.Unmarshal(data, set); err != nil {
		return nil, err
	}
	return set, nil
}

// Save persists the set, creating the enclosing directory if needed. An empty
// set removes the file so we don't leave an empty artifact behind.
func (self *Set) Save(worktreeGitDirPath string) error {
	path := storePath(worktreeGitDirPath)

	if !self.HasNamedChangelists() && self.Active == DefaultName {
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(self, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
