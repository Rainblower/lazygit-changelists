package changelists

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssignMovesPathBetweenChangelists(t *testing.T) {
	set := &Set{}
	set.Assign("a.go", "Feature")
	assert.Equal(t, "Feature", set.NameForPath("a.go"))

	// re-assigning moves the path rather than duplicating it
	set.Assign("a.go", "Bugfix")
	assert.Equal(t, "Bugfix", set.NameForPath("a.go"))
	assert.Empty(t, set.find("Feature").Paths)
	assert.Equal(t, []string{"a.go"}, set.find("Bugfix").Paths)
}

func TestAssignToDefaultRemovesFromNamedChangelists(t *testing.T) {
	set := &Set{}
	set.Assign("a.go", "Feature")
	set.Assign("a.go", DefaultName)
	assert.Equal(t, DefaultName, set.NameForPath("a.go"))
	assert.Empty(t, set.find("Feature").Paths)
}

func TestCreateRejectsDuplicatesAndDefault(t *testing.T) {
	set := &Set{}
	assert.True(t, set.Create("Feature"))
	assert.False(t, set.Create("Feature"))
	assert.False(t, set.Create(DefaultName))
	assert.Equal(t, []string{"Feature"}, set.Names())
}

func TestRenameMovesActivePointer(t *testing.T) {
	set := &Set{Active: "Feature"}
	set.Assign("a.go", "Feature")

	assert.True(t, set.Rename("Feature", "Renamed"))
	assert.Equal(t, "Renamed", set.Active)
	assert.Equal(t, "Renamed", set.NameForPath("a.go"))

	// can't rename onto an existing name
	set.Create("Other")
	assert.False(t, set.Rename("Renamed", "Other"))
}

func TestRemoveReturnsFilesToDefault(t *testing.T) {
	set := &Set{Active: "Feature"}
	set.Assign("a.go", "Feature")

	set.Remove("Feature")
	assert.Equal(t, DefaultName, set.NameForPath("a.go"))
	assert.Equal(t, DefaultName, set.Active)
	assert.False(t, set.HasNamedChangelists())
}

func TestRenamePathFollowsFileRename(t *testing.T) {
	set := &Set{}
	set.Assign("old.go", "Feature")

	assert.True(t, set.RenamePath("old.go", "new.go"))
	assert.Equal(t, "Feature", set.NameForPath("new.go"))
	assert.Equal(t, DefaultName, set.NameForPath("old.go"))

	// renaming an untracked path changes nothing
	assert.False(t, set.RenamePath("absent.go", "other.go"))
}

func TestPruneDropsStalePaths(t *testing.T) {
	set := &Set{}
	set.Assign("a.go", "Feature")
	set.Assign("b.go", "Feature")

	changed := set.Prune(map[string]bool{"a.go": true})
	assert.True(t, changed)
	assert.Equal(t, []string{"a.go"}, set.find("Feature").Paths)

	// nothing to prune the second time
	assert.False(t, set.Prune(map[string]bool{"a.go": true}))
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	set := &Set{Active: "Feature"}
	set.Assign("a.go", "Feature")
	set.Create("Empty")
	assert.NoError(t, set.Save(dir))

	loaded, err := Load(dir)
	assert.NoError(t, err)
	assert.Equal(t, "Feature", loaded.Active)
	assert.Equal(t, []string{"Feature", "Empty"}, loaded.Names())
	assert.Equal(t, "Feature", loaded.NameForPath("a.go"))
}

func TestLoadMissingFileYieldsEmptySet(t *testing.T) {
	set, err := Load(t.TempDir())
	assert.NoError(t, err)
	assert.False(t, set.HasNamedChangelists())
}

func TestSaveEmptySetRemovesFile(t *testing.T) {
	dir := t.TempDir()

	set := &Set{}
	set.Assign("a.go", "Feature")
	assert.NoError(t, set.Save(dir))

	set.Remove("Feature")
	assert.NoError(t, set.Save(dir))

	// an empty set should leave no file behind
	loaded, err := Load(dir)
	assert.NoError(t, err)
	assert.False(t, loaded.HasNamedChangelists())
}
