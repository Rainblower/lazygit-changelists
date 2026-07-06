package filetree

import (
	"sort"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/changelists"
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

// BuildChangelistGroupedTree lays the changed files out as a flat list ordered
// by changelist: the Default group (files not assigned to any named changelist)
// first, then each named changelist in the order they appear in the set. The
// ordering within a group is preserved from the flat build, and section headers
// for each group are added separately by the context's getNonModelItems.
func BuildChangelistGroupedTree(
	files []*models.File,
	showRootItem bool,
	cmp func(a, b *Node[models.File]) int,
	set *changelists.Set,
) *Node[models.File] {
	root := BuildFlatTreeFromFiles(files, showRootItem, cmp)

	groupOrder := map[string]int{}
	for i, name := range set.Names() {
		groupOrder[name] = i + 1
	}
	orderForNode := func(node *Node[models.File]) int {
		if node.File == nil {
			return 0
		}
		return groupOrder[set.NameForPath(node.File.Path)]
	}

	sort.SliceStable(root.Children, func(i, j int) bool {
		return orderForNode(root.Children[i]) < orderForNode(root.Children[j])
	})

	return root
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
