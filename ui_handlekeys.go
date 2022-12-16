package main

// This function handle every keyboards input in "view" mode
func (MyUI *UI) HandleKeyHelpMode(input string) string {
	switch input {
	// Quit
	case "q", "Q", "<C-c>", "<Escape>":
		MyUI.DisplayPopup("help", false)
	}
	return "continue"
}

// This function handle every keyboards input in "view" mode
func (MyUI *UI) HandleKeyViewMode(input string) string {
	// Initalize returnValue
	returnValue := "continue"

	// Handle keyboard inputs
	switch input {

	// Quit
	case "q", "<C-c>":
		return "quit"

	// Move in the list
	case "<Down>", "<MouseWheelDown>":
		if MyUI.List.Rows != nil {
			MyUI.List.ScrollDown()
		}
	case "<Up>", "<MouseWheelUp>":
		MyUI.List.ScrollUp()
	case "<PageDown>":
		if MyUI.List.Rows != nil {
			MyUI.List.ScrollPageDown()
		}
	case "<PageUp>":
		MyUI.List.ScrollPageUp()
	case "<Home>":
		MyUI.List.ScrollTop()
	case "<End>":
		if MyUI.List.Rows != nil {
			MyUI.List.ScrollBottom()
		}

	// Refresh VG List
	case "<C-r>", "<C-R>":
		MyUI.Log("VG list refresh requested", "yellow", "clear")
		GetVGList()
		MyUI.UpdateList()
		MyUI.Log("List updated", "yellow", "clear")

	// Select items
	case "s", "<Space>":
		MyUI.Select()

	// Select All
	case "<C-a>":
		MyUI.SelectAll()

	// Clear selection
	case "<C-<Space>>":
		MyUI.ClearSelection()

	// Change display mode
	case "d", "D":
		MyUI.ChangeDisplayMode()

	// Display help
	case "h", "H", "<F1>":
		MyUI.DisplayPopup("help", true)

	// Create filter
	case "f", "F", "/":
		MyUI.EnterFilter()
		MyUI.UpdateContentFilterZone(MyUI.Filter)
		MyUI.UpdateList()

	// Change order
	case "o", "O":
		MyUI.OrderList()

	// Ask for desc udpate
	case "U", "u":
		if MyUI.SelectedItems < 1 {
			MyUI.Select()
		}
		MyUI.UpdateVGDescription()
		MyUI.ClearSelection()

	// Delete filter
	case "<C-f>", "<Escape>":
		MyUI.Filter = ""
		MyUI.UpdateContentFilterZone("")
		MyUI.UpdateList()

	// Delete VG
	case "<C-d>":
		if MyUI.SelectedItems < 1 {
			MyUI.Select()
		}
		if MyUI.RequestDeleteVG() {
			GetVGList()
			MyUI.ClearSelection()
			MyUI.UpdateList()
			MyUI.Log("List updated", "yellow", "clear")
		}

	// In case of resize of the terminal
	case "<Resize>":
		MyUI.Resize()
		MyUI.UpdateList()
	}

	MyUI.UpdateDetail()
	MyUI.Render()

	// Return value
	return returnValue
}
