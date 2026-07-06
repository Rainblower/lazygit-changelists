package context

import (
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/changelists"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/icons"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

type WorkingTreeContext struct {
	*filetree.FileTreeViewModel
	*ListContextTrait
}

var (
	_ types.IListContext       = (*WorkingTreeContext)(nil)
	_ types.IFilterableContext = (*WorkingTreeContext)(nil)
)

func NewWorkingTreeContext(c *ContextCommon) *WorkingTreeContext {
	viewModel := filetree.NewFileTreeViewModel(
		func() []*models.File { return c.Model().Files },
		func() *changelists.Set { return c.Model().Changelists },
		c.Common,
		c.UserConfig().Gui.ShowFileTree,
	)

	getDisplayStrings := func(_ int, _ int) [][]string {
		showFileIcons := icons.IsIconEnabled() && c.UserConfig().Gui.ShowFileIcons
		showNumstat := c.UserConfig().Gui.ShowNumstatInFilesView
		lines := presentation.RenderFileTree(viewModel, c.Model().Submodules, showFileIcons, showNumstat, &c.UserConfig().Gui.CustomIcons, c.UserConfig().Gui.ShowRootItemInFileTree)
		return lo.Map(lines, func(line string, _ int) []string {
			return []string{line}
		})
	}

	// Insert a section header before the first file of each changelist group.
	// The tree lays files out grouped by changelist (see BuildChangelistGroupedTree),
	// so a group boundary is simply where the changelist of consecutive files
	// changes.
	getNonModelItems := func() []*NonModelItem {
		set := c.Model().Changelists
		if set == nil || !set.HasNamedChangelists() {
			return nil
		}

		result := []*NonModelItem{}
		prevName := ""
		started := false
		for i, node := range viewModel.GetAllItems() {
			if node.File == nil {
				continue
			}
			name := set.NameForPath(node.File.Path)
			if started && name == prevName {
				continue
			}
			result = append(result, &NonModelItem{
				Index:   i,
				Content: presentation.ChangelistHeaderLine(name, name == set.Active, c.Tr),
			})
			prevName = name
			started = true
		}
		return result
	}

	ctx := &WorkingTreeContext{
		FileTreeViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				View:       c.Views().Files,
				WindowName: "files",
				Key:        FILES_CONTEXT_KEY,
				Kind:       types.SIDE_CONTEXT,
				Focusable:  true,
			})),
			ListRenderer: ListRenderer{
				list:              viewModel,
				getDisplayStrings: getDisplayStrings,
				getNonModelItems:  getNonModelItems,
			},
			c: c,
		},
	}

	return ctx
}
