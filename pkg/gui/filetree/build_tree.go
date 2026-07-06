package filetree

import (
	"sort"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/changelists"
	"github.com/samber/lo"
)

func BuildTreeFromFiles(
	files []*models.File,
	showRootItem bool,
	cmp func(a, b *Node[models.File]) int,
) *Node[models.File] {
	root := &Node[models.File]{}

	childrenMapsByNode := make(map[*Node[models.File]]map[string]*Node[models.File])

	var curr *Node[models.File]
	for _, file := range files {
		splitPath := SplitFileTreePath(file.Path, showRootItem)
		curr = root
	outer:
		for i := range splitPath {
			var setFile *models.File
			isFile := i == len(splitPath)-1
			if isFile {
				setFile = file
			}

			path := join(splitPath[:i+1])

			var currNodeChildrenMap map[string]*Node[models.File]
			var isCurrNodeMapped bool

			if currNodeChildrenMap, isCurrNodeMapped = childrenMapsByNode[curr]; !isCurrNodeMapped {
				currNodeChildrenMap = make(map[string]*Node[models.File])
				childrenMapsByNode[curr] = currNodeChildrenMap
			}

			child, doesCurrNodeHaveChildAlready := currNodeChildrenMap[path]
			if doesCurrNodeHaveChildAlready {
				curr = child
				continue outer
			}

			if i == 0 && len(files) == 1 && len(splitPath) == 2 {
				// skip the root item when there's only one file at top level; we don't need it in that case
				continue outer
			}

			newChild := &Node[models.File]{
				path: path,
				File: setFile,
			}
			curr.Children = append(curr.Children, newChild)

			currNodeChildrenMap[path] = newChild

			curr = newChild
		}
	}

	root.Sort(cmp)
	root.Compress()

	return root
}

func BuildFlatTreeFromCommitFiles(
	files []*models.CommitFile,
	showRootItem bool,
	cmp func(a, b *Node[models.CommitFile]) int,
) *Node[models.CommitFile] {
	rootAux := BuildTreeFromCommitFiles(files, showRootItem, cmp)
	sortedFiles := rootAux.GetLeaves()

	return &Node[models.CommitFile]{Children: sortedFiles}
}

func BuildTreeFromCommitFiles(
	files []*models.CommitFile,
	showRootItem bool,
	cmp func(a, b *Node[models.CommitFile]) int,
) *Node[models.CommitFile] {
	root := &Node[models.CommitFile]{}

	var curr *Node[models.CommitFile]
	for _, file := range files {
		splitPath := SplitFileTreePath(file.Path, showRootItem)
		curr = root
	outer:
		for i := range splitPath {
			var setFile *models.CommitFile
			isFile := i == len(splitPath)-1
			if isFile {
				setFile = file
			}

			path := join(splitPath[:i+1])

			for _, existingChild := range curr.Children {
				if existingChild.path == path {
					curr = existingChild
					continue outer
				}
			}

			if i == 0 && len(files) == 1 && len(splitPath) == 2 {
				// skip the root item when there's only one file at top level; we don't need it in that case
				continue outer
			}

			newChild := &Node[models.CommitFile]{
				path: path,
				File: setFile,
			}
			curr.Children = append(curr.Children, newChild)

			curr = newChild
		}
	}

	root.Sort(cmp)
	root.Compress()

	return root
}

func BuildFlatTreeFromFiles(
	files []*models.File,
	showRootItem bool,
	cmp func(a, b *Node[models.File]) int,
) *Node[models.File] {
	rootAux := BuildTreeFromFiles(files, showRootItem, cmp)
	sortedFiles := rootAux.GetLeaves()

	// from top down we have merge conflict files, then tracked file, then untracked
	// files. This is the one way in which sorting differs between flat mode and
	// tree mode
	sort.SliceStable(sortedFiles, func(i, j int) bool {
		iFile := sortedFiles[i].File
		jFile := sortedFiles[j].File

		// never going to happen but just to be safe
		if iFile == nil || jFile == nil {
			return false
		}

		if iFile.HasMergeConflicts && !jFile.HasMergeConflicts {
			return true
		}

		if jFile.HasMergeConflicts && !iFile.HasMergeConflicts {
			return false
		}

		if iFile.Tracked && !jFile.Tracked {
			return true
		}

		if jFile.Tracked && !iFile.Tracked {
			return false
		}

		return false
	})

	return &Node[models.File]{Children: sortedFiles}
}

// BuildChangelistGroupedTree lays the changed files out grouped by changelist:
// the Default group (files not assigned to any named changelist) first, then
// each named changelist in the order they appear in the set. Each group gets
// its own independently-built subtree, so directory structure and path
// compression work within a group exactly as they do without changelists; when
// showTree is false each group is a flat list instead. Section headers for each
// group are added separately by the context's getNonModelItems.
//
// The root item ("./") is never shown here: the changelist headers already
// provide the top-level grouping, so a per-group root item would just be noise.
func BuildChangelistGroupedTree(
	files []*models.File,
	cmp func(a, b *Node[models.File]) int,
	set *changelists.Set,
	showTree bool,
) *Node[models.File] {
	groupNames := append([]string{changelists.DefaultName}, set.Names()...)

	children := []*Node[models.File]{}
	for _, name := range groupNames {
		groupFiles := lo.Filter(files, func(file *models.File, _ int) bool {
			return set.NameForPath(file.Path) == name
		})
		if len(groupFiles) == 0 {
			continue
		}

		var groupRoot *Node[models.File]
		if showTree {
			groupRoot = BuildTreeFromFiles(groupFiles, false, cmp)
		} else {
			groupRoot = BuildFlatTreeFromFiles(groupFiles, false, cmp)
		}
		children = append(children, groupRoot.Children...)
	}

	return &Node[models.File]{Children: children}
}

func split(str string) []string {
	return strings.Split(str, "/")
}

func join(strs []string) string {
	return strings.Join(strs, "/")
}

func SplitFileTreePath(path string, showRootItem bool) []string {
	return split(InternalTreePathForFilePath(path, showRootItem))
}

func InternalTreePathForFilePath(path string, showRootItem bool) string {
	if showRootItem {
		return "./" + path
	}

	return path
}
