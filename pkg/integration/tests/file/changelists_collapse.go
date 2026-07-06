package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ChangelistsCollapse = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Collapse and expand a changelist group like a directory",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowFileTree = true
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("file1", "content1\n")
		shell.CreateFile("file2", "content2\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			// move file1 into a new changelist so we have two groups
			Press(keys.Files.ViewChangelistOptions).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Changelist options")).
					Select(Contains("Move selected file(s) to changelist")).
					Confirm()

				t.ExpectPopup().Menu().
					Title(Equals("Move to changelist")).
					Select(Contains("New changelist")).
					Confirm()

				t.ExpectPopup().Prompt().
					Title(Equals("Enter a name for the new changelist")).
					Type("Feature").
					Confirm()
			}).
			Lines(
				Contains("▼ Default"),
				Contains("file2"),
				Contains("▼ Feature"),
				Contains("file1"),
			).
			// collapsing hides each group's files, leaving just the headers
			Press(keys.Files.CollapseAll).
			Lines(
				Contains("▶ Default"),
				Contains("▶ Feature"),
			).
			// expanding brings the files back
			Press(keys.Files.ExpandAll).
			Lines(
				Contains("▼ Default"),
				Contains("file2"),
				Contains("▼ Feature"),
				Contains("file1"),
			)
	},
})
