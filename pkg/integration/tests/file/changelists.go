package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var Changelists = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Group changed files into a named changelist, then rename and delete it",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		// flat view so grouping is easy to assert on
		config.GetUserConfig().Gui.ShowFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("file1", "content1\n")
		shell.CreateFile("file2", "content2\n")
		shell.CreateFile("file3", "content3\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Equals("?? file1").IsSelected(),
				Equals("?? file2"),
				Equals("?? file3"),
			).
			// move the selected file into a brand-new changelist
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
			// file1 now sits under the Feature group; the others stay in Default
			Lines(
				Contains("Default"),
				Equals("?? file2"),
				Equals("?? file3"),
				Contains("Feature"),
				Equals("?? file1"),
			)

		// rename the changelist
		t.Views().Files().
			Press(keys.Files.ViewChangelistOptions).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Changelist options")).
					Select(Contains("Rename changelist")).
					Confirm()

				t.ExpectPopup().Menu().
					Title(Equals("Select changelist to rename")).
					Select(Equals("Feature")).
					Confirm()

				t.ExpectPopup().Prompt().
					Title(Contains("Rename changelist")).
					Clear().
					Type("Renamed").
					Confirm()
			}).
			Lines(
				Contains("Default"),
				Equals("?? file2"),
				Equals("?? file3"),
				Contains("Renamed"),
				Equals("?? file1"),
			)

		// delete the changelist; its file returns to Default and grouping goes away
		t.Views().Files().
			Press(keys.Files.ViewChangelistOptions).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Changelist options")).
					Select(Contains("Delete changelist")).
					Confirm()

				t.ExpectPopup().Menu().
					Title(Equals("Select changelist to delete")).
					Select(Equals("Renamed")).
					Confirm()
			}).
			Lines(
				Equals("?? file1"),
				Equals("?? file2"),
				Equals("?? file3"),
			)
	},
})
