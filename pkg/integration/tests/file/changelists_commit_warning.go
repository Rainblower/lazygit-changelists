package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ChangelistsCommitWarning = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Warn before committing staged files that span multiple changelists",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("file1", "content1\n")
		shell.CreateFile("file2", "content2\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			// move file1 into changelist "A"; file2 stays in Default
			NavigateToLine(Contains("file1")).
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
					Type("A").
					Confirm()
			}).
			// stage everything: now staged files span Default (file2) and A (file1)
			Press(keys.Files.ToggleStagedAll).
			// committing warns that the staged files span multiple changelists
			Press(keys.Files.CommitChanges).
			Tap(func() {
				t.ExpectPopup().Confirmation().
					Title(Equals("Commit spans multiple changelists")).
					Content(Contains("more than one changelist")).
					Confirm()

				// confirming proceeds to the normal commit message panel
				t.ExpectPopup().CommitMessagePanel().
					Type("my commit").
					Confirm()
			})

		// the commit went through: both files are gone, leaving only the
		// now-empty "A" changelist header (empty named changelists stay visible)
		t.Views().Files().Lines(
			Equals("A"),
		)
	},
})
