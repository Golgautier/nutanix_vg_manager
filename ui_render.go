package main

import (
	"regexp"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// Const for UI positioning
const (
	ConstTitleHeight        int    = 3
	ConstDetailHeight       int    = 11
	ConstInteractHeight     int    = 3
	ConstFooterHeight       int    = 3
	ConstFilterZoneShift    int    = 43
	ConstMaxANDFiltering    int    = 10
	ConstZoneActive         string = "(fg:black,bg:green)"
	ConstZonePassive        string = "(fg:black,bg:white)"
	ConstZoneError          string = "(fg:white,bg:red)"
	ConstSelectionHighlight string = "(fg:green,bg:clear)"
	ConstFilterHighlight    string = "(fg:black,bg:green)"
)

// define Filtering struct
type filter_struct struct {
	UUID        [ConstMaxANDFiltering]*regexp.Regexp
	Name        [ConstMaxANDFiltering]*regexp.Regexp
	Cluster     [ConstMaxANDFiltering]*regexp.Regexp
	Container   [ConstMaxANDFiltering]*regexp.Regexp
	Mounted     [ConstMaxANDFiltering]*regexp.Regexp
	Description [ConstMaxANDFiltering]*regexp.Regexp
	Size        [ConstMaxANDFiltering]*regexp.Regexp
	Categories  [ConstMaxANDFiltering]*regexp.Regexp
}

// Define Struct for UI design & content
type UI struct {
	Title         *widgets.Paragraph
	List          *widgets.List
	Detail        *widgets.Paragraph
	Interact      *widgets.Paragraph
	Footer        *widgets.Paragraph
	FilterZone    *widgets.Paragraph
	Debug         *widgets.Paragraph
	Popup         *widgets.Paragraph
	SelectedItems int
	DisplayOrder  string // uuid | name
	Mode          string // view | help
	TermWidth     int
	TermHeight    int
	Filter        string
	AdvFilter     filter_struct
	Order         string // asc | desc
	TotalVG       int
	UglyFix       bool
}

// Create/initialize UI
func (MyUI *UI) Create() {
	// Set default mode and order
	MyUI.Mode = "view"
	MyUI.Order = "asc"
	MyUI.SelectedItems = 0

	// Use a string that will match to nothing as intialization
	tmp, _ := regexp.Compile("###")

	// Intialize the filters
	for i := 0; i < ConstMaxANDFiltering; i++ {
		MyUI.AdvFilter.Container[i] = tmp
		MyUI.AdvFilter.UUID[i] = tmp
		MyUI.AdvFilter.Name[i] = tmp
		MyUI.AdvFilter.Cluster[i] = tmp
		MyUI.AdvFilter.Mounted[i] = tmp
		MyUI.AdvFilter.Description[i] = tmp
		MyUI.AdvFilter.Size[i] = tmp
		MyUI.AdvFilter.Categories[i] = tmp
	}

	// Get terminal size
	MyUI.TermWidth, MyUI.TermHeight = ui.TerminalDimensions()

	MyUI.Title = widgets.NewParagraph()
	MyUI.Title.Border = false
	MyUI.Title.Text = "Nutanix VG Manager"
	MyUI.Title.TextStyle.Fg = ui.ColorCyan
	MyUI.Title.PaddingLeft = (MyUI.TermWidth - 3 - len(MyUI.Title.Text)) / 2
	MyUI.Title.SetRect(0, 0, MyUI.TermWidth, ConstTitleHeight)

	MyUI.List = widgets.NewList()
	MyUI.List.Title = "VG List - Name (UUID) - Sorted : " + MyUI.Order
	MyUI.List.SelectedRowStyle.Bg = ui.ColorCyan
	MyUI.List.SelectedRowStyle.Fg = ui.ColorBlack
	MyUI.List.WrapText = false
	MyUI.List.PaddingLeft = 1
	MyUI.List.PaddingRight = 1
	MyUI.List.Rows = []string{}
	MyUI.List.SetRect(0, ConstTitleHeight, MyUI.TermWidth, MyUI.TermHeight-ConstDetailHeight-ConstInteractHeight-ConstFooterHeight+2)
	MyUI.DisplayOrder = "name"

	MyUI.Detail = widgets.NewParagraph()
	MyUI.Detail.Title = "Details"
	MyUI.Detail.WrapText = false
	MyUI.Detail.PaddingLeft = 1
	MyUI.Detail.PaddingTop = 1
	MyUI.Detail.SetRect(0, MyUI.TermHeight-ConstDetailHeight-ConstInteractHeight-ConstFooterHeight+2, MyUI.TermWidth, MyUI.TermHeight-ConstInteractHeight-ConstFooterHeight+2)

	MyUI.Interact = widgets.NewParagraph()
	MyUI.Interact.Title = "Logs/Actions"
	MyUI.Interact.WrapText = false
	MyUI.Interact.Border = true
	MyUI.Interact.PaddingLeft = 1
	MyUI.Interact.PaddingTop = 0
	MyUI.Interact.SetRect(0, MyUI.TermHeight-ConstInteractHeight-ConstFooterHeight+2, MyUI.TermWidth, MyUI.TermHeight-ConstFooterHeight+2)

	MyUI.Footer = widgets.NewParagraph()
	MyUI.Footer.Border = false
	MyUI.Footer.Text = "h : Help / Q : Quit"
	MyUI.Footer.SetRect(0, MyUI.TermHeight-ConstFooterHeight+1, MyUI.TermWidth, MyUI.TermHeight+1)
	MyUI.Footer.WrapText = true

	MyUI.FilterZone = widgets.NewParagraph()
	MyUI.FilterZone.Border = false
	MyUI.FilterZone.Text = "Filter : [                              ]" + ConstZonePassive
	MyUI.FilterZone.SetRect(MyUI.TermWidth-ConstFilterZoneShift, ConstTitleHeight, MyUI.TermWidth-2, ConstTitleHeight+1)

	MyUI.Debug = widgets.NewParagraph()
	MyUI.Debug.TextStyle.Bg = ui.ColorRed
	MyUI.Debug.Border = false
	MyUI.Debug.SetRect(MyUI.TermWidth-50, 0, MyUI.TermWidth, 3)

	MyUI.Popup = widgets.NewParagraph()
	MyUI.Popup.Border = true
	MyUI.Popup.PaddingLeft = 1
	MyUI.Popup.PaddingTop = 1
	MyUI.Popup.SetRect(0, 0, 0, 0) // We do not display it right now

	// Fill list and detail
	MyUI.UpdateList()
	MyUI.UpdateDetail()
}

// Function to resize UI elements after terminal size change
func (MyUI *UI) Resize() {

	MyUI.TermWidth, MyUI.TermHeight = ui.TerminalDimensions()
	MyUI.Title.PaddingLeft = (MyUI.TermWidth - 3 - len(MyUI.Title.Text)) / 2
	MyUI.Title.SetRect(0, 0, MyUI.TermWidth, ConstTitleHeight)
	MyUI.List.SetRect(0, ConstTitleHeight, MyUI.TermWidth, MyUI.TermHeight-ConstDetailHeight-ConstInteractHeight-ConstFooterHeight+2)
	MyUI.Detail.SetRect(0, MyUI.TermHeight-ConstDetailHeight-ConstInteractHeight-ConstFooterHeight+2, MyUI.TermWidth, MyUI.TermHeight-ConstInteractHeight-ConstFooterHeight+2)
	MyUI.Interact.SetRect(0, MyUI.TermHeight-ConstInteractHeight-ConstFooterHeight+2, MyUI.TermWidth, MyUI.TermHeight-ConstFooterHeight+2)
	MyUI.Footer.SetRect(0, MyUI.TermHeight-ConstFooterHeight+1, MyUI.TermWidth, MyUI.TermHeight+1)
	MyUI.FilterZone.SetRect(MyUI.TermWidth-ConstFilterZoneShift, ConstTitleHeight, MyUI.TermWidth-2, ConstTitleHeight+1)
	MyUI.Debug.SetRect(MyUI.TermWidth-50, 0, MyUI.TermWidth, 3)
	MyUI.Popup.SetRect(0, 0, 0, 0) // We do not display it right now
}

// Render UI
func (MyUI UI) Render() {
	ui.Render(MyUI.Title, MyUI.List, MyUI.Detail, MyUI.Footer, MyUI.Interact, MyUI.Debug, MyUI.FilterZone, MyUI.Popup)
}
