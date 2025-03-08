package main

import (
	"fmt"
	"log"
	ntnx "nvgm/ntnx_api_call"

	ui "github.com/gizak/termui/v3"
	tb "github.com/nsf/termbox-go"
)

// Global variable
var MyPrism ntnx.Ntnx_endpoint

func main() {
	// Var declaration
	var MyUI UI
	var action string

	// Get Prism info
	GetPrismInfo()

	// Test API connection
	fmt.Print("Test connection to PC...")
	MyPrism.CallAPIJSON("PC", "GET", "/api/nutanix/v3/users/me", "", nil)
	fmt.Println("Ok")

	fmt.Println("Please wait during VG List collection...")

	GetVGList()

	fmt.Println("Done")

	// Initialize TermUI engine
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// DisableMouseInput in termbox-go to allow copy/paste works
	tb.SetInputMode(tb.InputEsc)

	// Create initial UI
	MyUI.Create()
	MyUI.Resize()
	MyUI.Render()

	// 2 types of events : time & keyboards inputs
	//tick := time.NewTicker(1 * time.Second)
	uiEvents := ui.PollEvents()

	for action != "quit" {
		select {
		//case <-tick.C:
		//	MyUI.UpdateListTitle()

		case e := <-uiEvents:
			switch MyUI.Mode {
			case "view":
				action = MyUI.HandleKeyViewMode(e.ID)
			case "help":
				action = MyUI.HandleKeyHelpMode(e.ID)
			}

			// If MyUI.Ask action have been executed, we need to 'flush' the chan of uiEvents
			// because entry will be seen here and will cause problems.
			// Note : we set MyUI.uglyfix at 'true' when function ask is called
			if MyUI.UglyFix == true {
				<-uiEvents
				MyUI.UglyFix = false
			}

		}
		MyUI.Render()
	}
}
