package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ChangelistsDeleteEmptied = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Offer to delete a named changelist once committing empties it",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("file1", "content1\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			// put the only file into changelist "A"
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
			// stage and commit it (only one changelist, so no cross-changelist warning)
			Press(keys.Files.ToggleStagedAll).
			Press(keys.Files.CommitChanges).
			Tap(func() {
				t.ExpectPopup().CommitMessagePanel().
					Type("my commit").
					Confirm()

				// committing emptied changelist "A", so we're offered to delete it
				t.ExpectPopup().Confirmation().
					Title(Equals("Changelist empty")).
					Content(Contains("now empty")).
					Confirm()
			})

		// after deleting the empty changelist, the panel is clean
		t.Views().Files().IsEmpty()
	},
})
