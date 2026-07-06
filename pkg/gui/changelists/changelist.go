// Package changelists implements JetBrains-style changelists for the Files
// panel: named groups that partition the working tree's changed files.
//
// Git has no native concept of a changelist, so the grouping is a client-side
// overlay. We store a mapping of file path to changelist name in the worktree's
// git dir (see store.go); files that aren't assigned to any changelist belong
// to the implicit "Default" changelist, which is never persisted.
package changelists

import (
	"slices"

	"github.com/samber/lo"
)

// DefaultName is the implicit changelist that owns every file not assigned to a
// named changelist. It is never stored; a file belongs to it precisely when no
// named changelist lists its path.
const DefaultName = ""

// Changelist is a named group of file paths.
type Changelist struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

// Set is the full collection of named changelists for a worktree, plus the
// active changelist that newly-changed files are assigned to.
type Set struct {
	Changelists []*Changelist `json:"changelists"`
	Active      string        `json:"active"`
}

// HasNamedChangelists reports whether any named changelist exists. When false,
// the Files panel renders exactly as it did before changelists existed, so we
// don't clutter it with a lone "Default" header.
func (self *Set) HasNamedChangelists() bool {
	return len(self.Changelists) > 0
}

// Names returns the names of the named changelists, in stored order.
func (self *Set) Names() []string {
	return lo.Map(self.Changelists, func(cl *Changelist, _ int) string {
		return cl.Name
	})
}

func (self *Set) find(name string) *Changelist {
	for _, cl := range self.Changelists {
		if cl.Name == name {
			return cl
		}
	}
	return nil
}

// NameForPath returns the name of the changelist that owns the given path, or
// DefaultName if no named changelist claims it.
func (self *Set) NameForPath(path string) string {
	for _, cl := range self.Changelists {
		if slices.Contains(cl.Paths, path) {
			return cl.Name
		}
	}
	return DefaultName
}

// Create adds a new empty changelist. It returns false if a changelist with
// that name already exists or the name is invalid (empty).
func (self *Set) Create(name string) bool {
	if name == DefaultName || self.find(name) != nil {
		return false
	}
	self.Changelists = append(self.Changelists, &Changelist{Name: name})
	return true
}

// Assign moves a path into the named changelist, removing it from any other
// changelist first so that each path lives in exactly one place. Assigning to
// DefaultName just removes the path from every named changelist. The target
// changelist is created on demand if it doesn't exist yet.
func (self *Set) Assign(path string, name string) {
	self.removePath(path)
	if name == DefaultName {
		return
	}
	cl := self.find(name)
	if cl == nil {
		cl = &Changelist{Name: name}
		self.Changelists = append(self.Changelists, cl)
	}
	if !slices.Contains(cl.Paths, path) {
		cl.Paths = append(cl.Paths, path)
	}
}

func (self *Set) removePath(path string) {
	for _, cl := range self.Changelists {
		cl.Paths = lo.Without(cl.Paths, path)
	}
}

// Rename changes a changelist's name, moving the active pointer with it. It
// returns false if the source doesn't exist or the target name is already
// taken.
func (self *Set) Rename(oldName string, newName string) bool {
	if newName == DefaultName || self.find(newName) != nil {
		return false
	}
	cl := self.find(oldName)
	if cl == nil {
		return false
	}
	cl.Name = newName
	if self.Active == oldName {
		self.Active = newName
	}
	return true
}

// Remove deletes a changelist; its files fall back to the Default changelist.
// If the removed changelist was active, Default becomes active.
func (self *Set) Remove(name string) {
	self.Changelists = lo.Filter(self.Changelists, func(cl *Changelist, _ int) bool {
		return cl.Name != name
	})
	if self.Active == name {
		self.Active = DefaultName
	}
}

// RenamePath updates a path in place when a file is renamed, so its changelist
// membership follows the rename. It returns whether an entry was actually
// updated (false if the old path isn't tracked by any changelist).
func (self *Set) RenamePath(oldPath string, newPath string) bool {
	for _, cl := range self.Changelists {
		if idx := slices.Index(cl.Paths, oldPath); idx != -1 {
			cl.Paths[idx] = newPath
			return true
		}
	}
	return false
}

// Prune drops any path not present in the given set of live paths, and returns
// whether anything changed. This keeps the store from accumulating stale
// entries for files that were committed or discarded outside lazygit.
func (self *Set) Prune(livePaths map[string]bool) bool {
	changed := false
	for _, cl := range self.Changelists {
		kept := lo.Filter(cl.Paths, func(path string, _ int) bool {
			return livePaths[path]
		})
		if len(kept) != len(cl.Paths) {
			cl.Paths = kept
			changed = true
		}
	}
	return changed
}
