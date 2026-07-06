package filetree

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/changelists"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBuildTreeFromFiles(t *testing.T) {
	scenarios := []struct {
		name         string
		files        []*models.File
		showRootItem bool
		expected     *Node[models.File]
	}{
		{
			name:  "no files",
			files: []*models.File{},
			expected: &Node[models.File]{
				path:     "",
				Children: nil,
			},
		},
		{
			name: "files in same directory",
			files: []*models.File{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						path:             "./dir1",
						CompressionLevel: 1,
						Children: []*Node[models.File]{
							{
								File: &models.File{Path: "dir1/a"},
								path: "./dir1/a",
							},
							{
								File: &models.File{Path: "dir1/b"},
								path: "./dir1/b",
							},
						},
					},
				},
			},
		},
		{
			name: "files in same directory, not root item",
			files: []*models.File{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						path:             "dir1",
						CompressionLevel: 0,
						Children: []*Node[models.File]{
							{
								File: &models.File{Path: "dir1/a"},
								path: "dir1/a",
							},
							{
								File: &models.File{Path: "dir1/b"},
								path: "dir1/b",
							},
						},
					},
				},
			},
		},
		{
			name: "paths that can be compressed",
			files: []*models.File{
				{
					Path: "dir1/dir3/a",
				},
				{
					Path: "dir2/dir4/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						path: ".",
						Children: []*Node[models.File]{
							{
								path: "./dir1/dir3",
								Children: []*Node[models.File]{
									{
										File: &models.File{Path: "dir1/dir3/a"},
										path: "./dir1/dir3/a",
									},
								},
								CompressionLevel: 1,
							},
							{
								path: "./dir2/dir4",
								Children: []*Node[models.File]{
									{
										File: &models.File{Path: "dir2/dir4/b"},
										path: "./dir2/dir4/b",
									},
								},
								CompressionLevel: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "paths that can be compressed, no root item",
			files: []*models.File{
				{
					Path: "dir1/dir3/a",
				},
				{
					Path: "dir2/dir4/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						path: "dir1/dir3",
						Children: []*Node[models.File]{
							{
								File: &models.File{Path: "dir1/dir3/a"},
								path: "dir1/dir3/a",
							},
						},
						CompressionLevel: 1,
					},
					{
						path: "dir2/dir4",
						Children: []*Node[models.File]{
							{
								File: &models.File{Path: "dir2/dir4/b"},
								path: "dir2/dir4/b",
							},
						},
						CompressionLevel: 1,
					},
				},
			},
		},
		{
			name: "paths that can be sorted",
			files: []*models.File{
				{
					Path: "b",
				},
				{
					Path: "a",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						path: ".",
						Children: []*Node[models.File]{
							{
								File: &models.File{Path: "a"},
								path: "./a",
							},
							{
								File: &models.File{Path: "b"},
								path: "./b",
							},
						},
					},
				},
			},
		},
		{
			name: "paths that can be sorted including a merge conflict file",
			files: []*models.File{
				{
					Path: "b",
				},
				{
					Path:              "z",
					HasMergeConflicts: true,
				},
				{
					Path: "a",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						path: ".",
						// it is a little strange that we're not bubbling up our merge conflict
						// here but we are technically still in tree mode and that's the rule
						Children: []*Node[models.File]{
							{
								File: &models.File{Path: "a"},
								path: "./a",
							},
							{
								File: &models.File{Path: "b"},
								path: "./b",
							},
							{
								File: &models.File{Path: "z", HasMergeConflicts: true},
								path: "./z",
							},
						},
					},
				},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			result := BuildTreeFromFiles(s.files, s.showRootItem, NodeSortComparator[models.File]("mixed", false))
			assert.EqualValues(t, s.expected, result)
		})
	}
}

func TestBuildFlatTreeFromFiles(t *testing.T) {
	scenarios := []struct {
		name         string
		files        []*models.File
		showRootItem bool
		expected     *Node[models.File]
	}{
		{
			name:  "no files",
			files: []*models.File{},
			expected: &Node[models.File]{
				path:     "",
				Children: []*Node[models.File]{},
			},
		},
		{
			name: "files in same directory",
			files: []*models.File{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						File:             &models.File{Path: "dir1/a"},
						path:             "./dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.File{Path: "dir1/b"},
						path:             "./dir1/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "files in same directory, not root item",
			files: []*models.File{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						File:             &models.File{Path: "dir1/a"},
						path:             "dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.File{Path: "dir1/b"},
						path:             "dir1/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "paths that can be compressed",
			files: []*models.File{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir2/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						File:             &models.File{Path: "dir1/a"},
						path:             "./dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.File{Path: "dir2/b"},
						path:             "./dir2/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "paths that can be compressed, no root item",
			files: []*models.File{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir2/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						File:             &models.File{Path: "dir1/a"},
						path:             "dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.File{Path: "dir2/b"},
						path:             "dir2/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "paths that can be sorted",
			files: []*models.File{
				{
					Path: "b",
				},
				{
					Path: "a",
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						File: &models.File{Path: "a"},
						path: "./a",
					},
					{
						File: &models.File{Path: "b"},
						path: "./b",
					},
				},
			},
		},
		{
			name: "tracked, untracked, and conflicted files",
			files: []*models.File{
				{
					Path:    "a2",
					Tracked: false,
				},
				{
					Path:    "a1",
					Tracked: false,
				},
				{
					Path:              "c2",
					HasMergeConflicts: true,
				},
				{
					Path:              "c1",
					HasMergeConflicts: true,
				},
				{
					Path:    "b2",
					Tracked: true,
				},
				{
					Path:    "b1",
					Tracked: true,
				},
			},
			showRootItem: true,
			expected: &Node[models.File]{
				path: "",
				Children: []*Node[models.File]{
					{
						File: &models.File{Path: "c1", HasMergeConflicts: true},
						path: "./c1",
					},
					{
						File: &models.File{Path: "c2", HasMergeConflicts: true},
						path: "./c2",
					},
					{
						File: &models.File{Path: "b1", Tracked: true},
						path: "./b1",
					},
					{
						File: &models.File{Path: "b2", Tracked: true},
						path: "./b2",
					},
					{
						File: &models.File{Path: "a1", Tracked: false},
						path: "./a1",
					},
					{
						File: &models.File{Path: "a2", Tracked: false},
						path: "./a2",
					},
				},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			result := BuildFlatTreeFromFiles(s.files, s.showRootItem, NodeSortComparator[models.File]("mixed", false))
			assert.EqualValues(t, s.expected, result)
		})
	}
}

func TestBuildTreeFromCommitFiles(t *testing.T) {
	scenarios := []struct {
		name         string
		files        []*models.CommitFile
		showRootItem bool
		expected     *Node[models.CommitFile]
	}{
		{
			name:  "no files",
			files: []*models.CommitFile{},
			expected: &Node[models.CommitFile]{
				path:     "",
				Children: nil,
			},
		},
		{
			name: "files in same directory",
			files: []*models.CommitFile{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						path:             "./dir1",
						CompressionLevel: 1,
						Children: []*Node[models.CommitFile]{
							{
								File: &models.CommitFile{Path: "dir1/a"},
								path: "./dir1/a",
							},
							{
								File: &models.CommitFile{Path: "dir1/b"},
								path: "./dir1/b",
							},
						},
					},
				},
			},
		},
		{
			name: "files in same directory, not root item",
			files: []*models.CommitFile{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						path:             "dir1",
						CompressionLevel: 0,
						Children: []*Node[models.CommitFile]{
							{
								File: &models.CommitFile{Path: "dir1/a"},
								path: "dir1/a",
							},
							{
								File: &models.CommitFile{Path: "dir1/b"},
								path: "dir1/b",
							},
						},
					},
				},
			},
		},
		{
			name: "paths that can be compressed",
			files: []*models.CommitFile{
				{
					Path: "dir1/dir3/a",
				},
				{
					Path: "dir2/dir4/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						path: ".",
						Children: []*Node[models.CommitFile]{
							{
								path: "./dir1/dir3",
								Children: []*Node[models.CommitFile]{
									{
										File: &models.CommitFile{Path: "dir1/dir3/a"},
										path: "./dir1/dir3/a",
									},
								},
								CompressionLevel: 1,
							},
							{
								path: "./dir2/dir4",
								Children: []*Node[models.CommitFile]{
									{
										File: &models.CommitFile{Path: "dir2/dir4/b"},
										path: "./dir2/dir4/b",
									},
								},
								CompressionLevel: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "paths that can be compressed, no root item",
			files: []*models.CommitFile{
				{
					Path: "dir1/dir3/a",
				},
				{
					Path: "dir2/dir4/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						path: "dir1/dir3",
						Children: []*Node[models.CommitFile]{
							{
								File: &models.CommitFile{Path: "dir1/dir3/a"},
								path: "dir1/dir3/a",
							},
						},
						CompressionLevel: 1,
					},
					{
						path: "dir2/dir4",
						Children: []*Node[models.CommitFile]{
							{
								File: &models.CommitFile{Path: "dir2/dir4/b"},
								path: "dir2/dir4/b",
							},
						},
						CompressionLevel: 1,
					},
				},
			},
		},
		{
			name: "paths that can be sorted",
			files: []*models.CommitFile{
				{
					Path: "b",
				},
				{
					Path: "a",
				},
			},
			showRootItem: true,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						path: ".",
						Children: []*Node[models.CommitFile]{
							{
								File: &models.CommitFile{Path: "a"},
								path: "./a",
							},
							{
								File: &models.CommitFile{Path: "b"},
								path: "./b",
							},
						},
					},
				},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			result := BuildTreeFromCommitFiles(s.files, s.showRootItem, NodeSortComparator[models.CommitFile]("mixed", false))
			assert.EqualValues(t, s.expected, result)
		})
	}
}

func TestBuildFlatTreeFromCommitFiles(t *testing.T) {
	scenarios := []struct {
		name         string
		files        []*models.CommitFile
		showRootItem bool
		expected     *Node[models.CommitFile]
	}{
		{
			name:  "no files",
			files: []*models.CommitFile{},
			expected: &Node[models.CommitFile]{
				path:     "",
				Children: []*Node[models.CommitFile]{},
			},
		},
		{
			name: "files in same directory",
			files: []*models.CommitFile{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						File:             &models.CommitFile{Path: "dir1/a"},
						path:             "./dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.CommitFile{Path: "dir1/b"},
						path:             "./dir1/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "files in same directory, not root item",
			files: []*models.CommitFile{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir1/b",
				},
			},
			showRootItem: false,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						File:             &models.CommitFile{Path: "dir1/a"},
						path:             "dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.CommitFile{Path: "dir1/b"},
						path:             "dir1/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "paths that can be compressed",
			files: []*models.CommitFile{
				{
					Path: "dir1/a",
				},
				{
					Path: "dir2/b",
				},
			},
			showRootItem: true,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						File:             &models.CommitFile{Path: "dir1/a"},
						path:             "./dir1/a",
						CompressionLevel: 0,
					},
					{
						File:             &models.CommitFile{Path: "dir2/b"},
						path:             "./dir2/b",
						CompressionLevel: 0,
					},
				},
			},
		},
		{
			name: "paths that can be sorted",
			files: []*models.CommitFile{
				{
					Path: "b",
				},
				{
					Path: "a",
				},
			},
			showRootItem: true,
			expected: &Node[models.CommitFile]{
				path: "",
				Children: []*Node[models.CommitFile]{
					{
						File: &models.CommitFile{Path: "a"},
						path: "./a",
					},
					{
						File: &models.CommitFile{Path: "b"},
						path: "./b",
					},
				},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			result := BuildFlatTreeFromCommitFiles(s.files, s.showRootItem, NodeSortComparator[models.CommitFile]("mixed", false))
			assert.EqualValues(t, s.expected, result)
		})
	}
}

func TestBuildChangelistGroupedTreeFlat(t *testing.T) {
	files := []*models.File{
		{Path: "a.go", Tracked: true},
		{Path: "b.go", Tracked: true},
		{Path: "c.go", Tracked: true},
		{Path: "d.go", Tracked: true},
	}
	set := &changelists.Set{}
	set.Assign("b.go", "Feature")
	set.Assign("d.go", "Feature")

	result := BuildChangelistGroupedTree(files, NodeSortComparator[models.File]("mixed", false), set, false)

	// two header nodes: Default first, then Feature
	assert.Len(t, result.Children, 2)
	assert.Equal(t, ChangelistNodePath(changelists.DefaultName), result.Children[0].GetInternalPath())
	assert.Equal(t, ChangelistNodePath("Feature"), result.Children[1].GetInternalPath())

	groupPaths := func(header *Node[models.File]) []string {
		return lo.Map(header.Children, func(node *Node[models.File], _ int) string {
			return node.File.Path
		})
	}

	// each group is a flat list of its own files
	assert.Equal(t, []string{"a.go", "c.go"}, groupPaths(result.Children[0]))
	assert.Equal(t, []string{"b.go", "d.go"}, groupPaths(result.Children[1]))
}

func TestBuildChangelistGroupedTreeShowsEmptyNamedChangelist(t *testing.T) {
	files := []*models.File{{Path: "a.go", Tracked: true}}
	set := &changelists.Set{}
	set.Create("Empty")

	result := BuildChangelistGroupedTree(files, NodeSortComparator[models.File]("mixed", false), set, false)

	// Default (with a.go) plus the empty named changelist, which is shown even
	// though it has no files so it can be seen and collapsed
	assert.Len(t, result.Children, 2)
	assert.Equal(t, ChangelistNodePath(changelists.DefaultName), result.Children[0].GetInternalPath())
	assert.Equal(t, ChangelistNodePath("Empty"), result.Children[1].GetInternalPath())
	assert.Empty(t, result.Children[1].Children)
}

func TestBuildChangelistGroupedTreeWithTree(t *testing.T) {
	files := []*models.File{
		{Path: "dir/a.go", Tracked: true},
		{Path: "dir/b.go", Tracked: true},
		{Path: "other/c.go", Tracked: true},
	}
	set := &changelists.Set{}
	set.Assign("dir/b.go", "Feature")

	result := BuildChangelistGroupedTree(files, NodeSortComparator[models.File]("mixed", false), set, true)

	// two header nodes, each holding its own directory tree
	assert.Len(t, result.Children, 2)
	defaultHeader := result.Children[0]
	featureHeader := result.Children[1]
	assert.Equal(t, ChangelistNodePath(changelists.DefaultName), defaultHeader.GetInternalPath())
	assert.Equal(t, ChangelistNodePath("Feature"), featureHeader.GetInternalPath())

	// Default group keeps the "dir" and "other" directories (a real tree);
	// Feature has its own "dir" holding just b.go, so "dir" appears in both
	// groups independently.
	defaultTop := lo.Map(defaultHeader.Children, func(node *Node[models.File], _ int) string {
		return node.GetPath()
	})
	assert.Equal(t, []string{"dir", "other"}, defaultTop)

	// the Default group's "dir" directory contains a.go
	assert.Nil(t, defaultHeader.Children[0].File)
	assert.Equal(t, "dir/a.go", defaultHeader.Children[0].Children[0].GetPath())

	assert.Equal(t, "dir/b.go", featureHeader.Children[0].GetPath())
}
