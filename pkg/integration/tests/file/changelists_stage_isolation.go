package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ChangelistsStageIsolation = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Staging a changelist only stages its own files, even when another changelist has a file in the same directory",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowFileTree = true
		config.GetUserConfig().Gui.ShowRootItemInFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		// three files in the same directory; two will go to changelist "A",
		// one stays in Default. "A" then holds a real "src" directory node
		// (not a compressed single-file path), which is where the bug bit.
		shell.CreateFile("src/a.go", "a\n")
		shell.CreateFile("src/b.go", "b\n")
		shell.CreateFile("src/c.go", "c\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		moveToNewChangelist := func(name string) {
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
				Type(name).
				Confirm()
		}

		t.Views().Files().
			IsFocused().
			NavigateToLine(Contains("a.go")).
			Press(keys.Files.ViewChangelistOptions).
			Tap(func() { moveToNewChangelist("A") }).
			// now move c.go into the existing "A" changelist too
			NavigateToLine(Contains("c.go")).
			Press(keys.Files.ViewChangelistOptions).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Changelist options")).
					Select(Contains("Move selected file(s) to changelist")).
					Confirm()
				t.ExpectPopup().Menu().
					Title(Equals("Move to changelist")).
					Select(Contains("A")).
					Confirm()
			}).
			Lines(
				Contains("▼ Default"),
				Contains("?? src/b.go"),
				Contains("▼ A"),
				Contains("src"),
				Contains("?? a.go"),
				Contains("?? c.go"),
			).
			// stage changelist A by pressing space on its header
			NavigateToLine(Contains("▼ A")).
			Press(keys.Universal.Select).
			Lines(
				Contains("▼ Default"),
				// b.go is in Default and must stay unstaged even though it lives
				// in the same "src" directory as the staged files of A
				Contains("?? src/b.go"),
				Contains("▼ A"),
				Contains("src"),
				Contains("A  a.go"),
				Contains("A  c.go"),
			).
			// unstage everything again by pressing space on the header
			NavigateToLine(Contains("▼ A")).
			Press(keys.Universal.Select).
			Lines(
				Contains("▼ Default"),
				Contains("?? src/b.go"),
				Contains("▼ A"),
				Contains("src"),
				Contains("?? a.go"),
				Contains("?? c.go"),
			).
			// staging the "src" directory node *inside* A must also stay isolated:
			// only A's src files get staged, not Default's b.go in the same dir
			NavigateToLine(Contains("▼ src")).
			Press(keys.Universal.Select).
			Lines(
				Contains("▼ Default"),
				Contains("?? src/b.go"),
				Contains("▼ A"),
				Contains("src"),
				Contains("A  a.go"),
				Contains("A  c.go"),
			)
	},
})
