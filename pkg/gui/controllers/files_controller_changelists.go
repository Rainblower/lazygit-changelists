package controllers

import (
	"errors"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/changelists"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

// openChangelistMenu is the single entry point for all changelist operations.
// Which items appear depends on context: moving files needs a selection, and
// renaming/deleting/activating needs at least one named changelist to exist.
func (self *FilesController) openChangelistMenu() error {
	set := self.changelists()

	menuItems := []*types.MenuItem{
		{
			Label:   self.c.Tr.NewChangelistMenuItem,
			OnPress: self.promptToCreateChangelist,
			Keys:    menuKey('n'),
		},
	}

	if selectedNodes, _, _ := self.context().GetSelectedItems(); len(selectedNodes) > 0 {
		menuItems = append(menuItems, &types.MenuItem{
			Label:   self.c.Tr.MoveToChangelistMenuItem,
			OnPress: func() error { return self.promptToMoveToChangelist(selectedNodes) },
			Keys:    menuKey('m'),
		})
	}

	if set.HasNamedChangelists() {
		menuItems = append(menuItems,
			&types.MenuItem{
				Label:   self.c.Tr.SetActiveChangelistMenuItem,
				OnPress: self.promptToSetActiveChangelist,
				Keys:    menuKey('s'),
			},
			&types.MenuItem{
				Label:   self.c.Tr.RenameChangelistMenuItem,
				OnPress: self.promptToRenameChangelist,
				Keys:    menuKey('r'),
			},
			&types.MenuItem{
				Label:   self.c.Tr.DeleteChangelistMenuItem,
				OnPress: self.promptToDeleteChangelist,
				Keys:    menuKey('d'),
			},
		)
	}

	return self.c.Menu(types.CreateMenuOptions{
		Title: self.c.Tr.ChangelistOptions,
		Items: menuItems,
	})
}

func (self *FilesController) promptToCreateChangelist() error {
	self.c.Prompt(types.PromptOpts{
		Title: self.c.Tr.NewChangelistPrompt,
		HandleConfirm: func(name string) error {
			if !self.changelists().Create(name) {
				return self.changelistNameError(name)
			}
			self.changelists().Active = name
			return self.saveAndRerenderChangelists()
		},
	})
	return nil
}

func (self *FilesController) promptToMoveToChangelist(selectedNodes []*filetree.FileNode) error {
	paths := filePathsOfNodes(selectedNodes)
	if len(paths) == 0 {
		return nil
	}

	assignTo := func(name string) error {
		for _, path := range paths {
			self.changelists().Assign(path, name)
		}
		return self.saveAndRerenderChangelists()
	}

	menuItems := []*types.MenuItem{
		{
			Label:   self.c.Tr.NewChangelistMenuItem,
			OnPress: func() error { return self.promptForNewChangelistThen(assignTo) },
			Keys:    menuKey('n'),
		},
		{
			Label:   self.c.Tr.MoveToDefaultMenuItem,
			OnPress: func() error { return assignTo(changelists.DefaultName) },
		},
	}
	for _, name := range self.changelists().Names() {
		menuItems = append(menuItems, &types.MenuItem{
			Label:   name,
			OnPress: func() error { return assignTo(name) },
		})
	}

	return self.c.Menu(types.CreateMenuOptions{
		Title: self.c.Tr.MoveToChangelistTitle,
		Items: menuItems,
	})
}

func (self *FilesController) promptToSetActiveChangelist() error {
	return self.chooseNamedChangelist(self.c.Tr.SetActiveChangelistTitle, func(name string) error {
		self.changelists().Active = name
		return self.saveAndRerenderChangelists()
	})
}

func (self *FilesController) promptToRenameChangelist() error {
	return self.chooseNamedChangelist(self.c.Tr.RenameChangelistTitle, func(name string) error {
		self.c.Prompt(types.PromptOpts{
			Title:          utils.ResolvePlaceholderString(self.c.Tr.RenameChangelistPrompt, map[string]string{"name": name}),
			InitialContent: name,
			HandleConfirm: func(newName string) error {
				if !self.changelists().Rename(name, newName) {
					return self.changelistNameError(newName)
				}
				return self.saveAndRerenderChangelists()
			},
		})
		return nil
	})
}

func (self *FilesController) promptToDeleteChangelist() error {
	return self.chooseNamedChangelist(self.c.Tr.DeleteChangelistTitle, func(name string) error {
		self.changelists().Remove(name)
		return self.saveAndRerenderChangelists()
	})
}

// chooseNamedChangelist shows a menu of the existing named changelists and
// invokes onChoose with the selected one.
func (self *FilesController) chooseNamedChangelist(title string, onChoose func(name string) error) error {
	menuItems := make([]*types.MenuItem, 0, len(self.changelists().Names()))
	for _, name := range self.changelists().Names() {
		menuItems = append(menuItems, &types.MenuItem{
			Label:   name,
			OnPress: func() error { return onChoose(name) },
		})
	}

	return self.c.Menu(types.CreateMenuOptions{Title: title, Items: menuItems})
}

func (self *FilesController) promptForNewChangelistThen(then func(name string) error) error {
	self.c.Prompt(types.PromptOpts{
		Title: self.c.Tr.NewChangelistPrompt,
		HandleConfirm: func(name string) error {
			if !self.changelists().Create(name) {
				return self.changelistNameError(name)
			}
			return then(name)
		},
	})
	return nil
}

func (self *FilesController) changelistNameError(name string) error {
	return errors.New(utils.ResolvePlaceholderString(
		self.c.Tr.InvalidChangelistName,
		map[string]string{"name": name},
	))
}

// changelists returns the model's changelist set, initialising it if the first
// files refresh hasn't populated it yet.
func (self *FilesController) changelists() *changelists.Set {
	if self.c.Model().Changelists == nil {
		self.c.Model().Changelists = &changelists.Set{}
	}
	return self.c.Model().Changelists
}

func (self *FilesController) saveAndRerenderChangelists() error {
	if err := self.changelists().Save(self.c.Git().RepoPaths.WorktreeGitDirPath()); err != nil {
		return err
	}
	self.context().FileTreeViewModel.SetTree()
	self.c.PostRefreshUpdate(self.context())
	return nil
}

// commitWithChangelistCheck wraps a commit handler so that if the staged files
// span more than one changelist (Default counts as one), the user is warned and
// asked to confirm before the commit proceeds. A commit is meant to correspond
// to a single changelist, so mixing them is usually a mistake — but we only warn
// (not block), since sometimes it's intentional.
func (self *FilesController) commitWithChangelistCheck(commitFn func() error) func() error {
	return func() error {
		if len(self.stagedChangelistNames()) <= 1 {
			return commitFn()
		}

		self.c.Confirm(types.ConfirmOpts{
			Title:         self.c.Tr.CommitSpansChangelistsTitle,
			Prompt:        self.c.Tr.CommitSpansChangelistsPrompt,
			HandleConfirm: commitFn,
		})
		return nil
	}
}

// stagedChangelistNames returns the distinct changelists (by name, with Default
// as "") that the currently staged files belong to. Returns nil when no named
// changelists exist, so the check is a no-op unless the feature is in use.
func (self *FilesController) stagedChangelistNames() []string {
	set := self.c.Model().Changelists
	if set == nil || !set.HasNamedChangelists() {
		return nil
	}

	seen := map[string]bool{}
	for _, file := range self.c.Model().Files {
		if file.HasStagedChanges {
			seen[set.NameForPath(file.Path)] = true
		}
	}

	return lo.Keys(seen)
}

// notOnChangelistHeader disables file-specific actions (open, diff, ignore,
// enter, …) when the selection is a changelist header, which has no file of its
// own.
func (self *FilesController) notOnChangelistHeader() *types.DisabledReason {
	node := self.getSelectedItem()
	if node != nil && node.IsChangelistHeader() {
		return &types.DisabledReason{Text: self.c.Tr.ChangelistHeaderNoFileAction}
	}
	return nil
}

func filePathsOfNodes(nodes []*filetree.FileNode) []string {
	paths := []string{}
	for _, node := range normalisedSelectedNodes(nodes) {
		_ = node.ForEachFile(func(file *models.File) error {
			paths = append(paths, file.Path)
			return nil
		})
	}
	return paths
}
