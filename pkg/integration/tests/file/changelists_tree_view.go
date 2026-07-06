package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ChangelistsTreeView = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Files keep their directory tree structure within a changelist group",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowFileTree = true
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("src/a.go", "a\n")
		shell.CreateFile("src/b.go", "b\n")
		shell.CreateFile("docs/c.md", "c\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Contains("docs").IsSelected(),
				Contains("c.md"),
				Contains("src"),
				Contains("a.go"),
				Contains("b.go"),
			).
			// move the whole docs directory into a new changelist
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
					Type("Docs").
					Confirm()
			}).
			// each group keeps its own directory tree: the src/ directory stays
			// in Default with its two files, docs/c.md moves under the Docs group
			Lines(
				Contains("Default"),
				Contains("src"),
				Contains("a.go"),
				Contains("b.go"),
				Contains("Docs"),
				Contains("docs/c.md"),
			)
	},
})
