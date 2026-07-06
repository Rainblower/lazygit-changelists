package filetree

import "strings"

// changelistNodePathPrefix marks a tree node as a changelist header rather than
// a real file or directory. It starts with a NUL byte, which can never appear
// in a real file path, so a changelist node's synthetic path never collides
// with a real one (and sorts before every real path). The changelist name
// follows the prefix, so the Default changelist's node path is exactly the
// prefix.
const changelistNodePathPrefix = "\x00changelist:"

// ChangelistNodePath is the synthetic internal path used for a changelist's
// header node in the tree.
func ChangelistNodePath(name string) string {
	return changelistNodePathPrefix + name
}

// IsChangelistHeader reports whether this node is a changelist header rather
// than a real file or directory.
func (self *FileNode) IsChangelistHeader() bool {
	return strings.HasPrefix(self.GetInternalPath(), changelistNodePathPrefix)
}

// ChangelistName returns the changelist name for a header node (empty string
// for the Default changelist). Only meaningful when IsChangelistHeader is true.
func (self *FileNode) ChangelistName() string {
	return strings.TrimPrefix(self.GetInternalPath(), changelistNodePathPrefix)
}
