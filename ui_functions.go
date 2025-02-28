package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	ui "github.com/gizak/termui/v3"
)

// Create map to identify if VG is Selected
var VGSelected = make(map[string]bool)

// Create pairing info between VG displayed and VG_List
// The index will match with index of list.Rows, the value will point on VG_List
var DisplayedVGPairing = []int{}

// Display line fonction
// Will handle selection and filter highlighting. Parameter : row of the VG_list to display
func (MyUI UI) DisplayLine(row int) string {
	// Create line
	var retvalue, prefix, suffix string
	var selected bool

	// Check if selected or not
	if VGSelected[GlobalVGList[row].UUID] {
		prefix = " * ["
		suffix = "]" + ConstSelectionHighlight
		selected = true
	} else {
		prefix = "   "
		suffix = ""
		selected = false
	}

	tmp := ""

	// Create viewable part regarding expected order
	if selected {
		// We do not add filter highlighting (not supported by TermUI)
		if MyUI.DisplayOrder == "uuid" {
			tmp = fmt.Sprintf("%s (%s)", GlobalVGList[row].UUID, GlobalVGList[row].Name)
		} else {
			tmp = fmt.Sprintf("%s (%s)", GlobalVGList[row].Name, GlobalVGList[row].UUID)
		}
	} else {
		if MyUI.DisplayOrder == "uuid" {
			tmp = fmt.Sprintf("%s (%s)", MyUI.AddFilterHighlight(GlobalVGList[row].UUID, MyUI.AdvFilter.UUID), MyUI.AddFilterHighlight(GlobalVGList[row].Name, MyUI.AdvFilter.Name))
		} else {
			tmp = fmt.Sprintf("%s (%s)", MyUI.AddFilterHighlight(GlobalVGList[row].Name, MyUI.AdvFilter.Name), MyUI.AddFilterHighlight(GlobalVGList[row].UUID, MyUI.AdvFilter.UUID))
		}
	}

	// Add space to get good length
	// (only if term_width > len tmp)
	d_length := len(GlobalVGList[row].Name) + len(GlobalVGList[row].UUID) + 10 // Count line chars, borders(4) and selection prefix(3)

	if d_length < MyUI.TermWidth {
		retvalue = fmt.Sprintf("%s%s%s%s", prefix, tmp, strings.Repeat(" ", MyUI.TermWidth-d_length), suffix)
	} else {
		retvalue = fmt.Sprintf("%s%s%s", prefix, tmp, suffix)
	}

	return retvalue
}

// Function to convert categories map to a string for matching
func categoriesMatchString(categories map[string]string) string {
	if len(categories) == 0 {
		return ""
	}
	
	var result strings.Builder
	for k, v := range categories {
		result.WriteString(fmt.Sprintf("%s:%s;", k, v))
	}
	return result.String()
}

// This function return true if vg match with filters
func (MyUI UI) MatchFilters(vg VG) bool {

	var num_filters int = strings.Count(MyUI.Filter, "&") + 1
	var results [ConstMaxANDFiltering]bool

	// We ensure not to go over filter limit
	if num_filters > ConstMaxANDFiltering {
		num_filters = ConstMaxANDFiltering
	}

	// Get categories as string for matching
	categoriesStr := categoriesMatchString(vg.Categories)

	// We check each filter one by one
	for i := 0; i < num_filters; i++ {

		// initialize result for this filter to false
		results[i] = false

		// We create a "CompareString" to be able to
		if MyUI.AdvFilter.Cluster[i].Match([]byte(vg.Cluster)) ||
			MyUI.AdvFilter.Name[i].Match([]byte(vg.Name)) ||
			MyUI.AdvFilter.Container[i].Match([]byte(vg.Container)) ||
			MyUI.AdvFilter.Size[i].Match([]byte(vg.Size)) ||
			MyUI.AdvFilter.UUID[i].Match([]byte(vg.UUID)) ||
			MyUI.AdvFilter.Mounted[i].Match([]byte(vg.Attached)) ||
			MyUI.AdvFilter.Description[i].Match([]byte(vg.Description)) ||
			MyUI.AdvFilter.Categories[i].Match([]byte(categoriesStr)) {

			// If one of the filters is ok, we store "true" state
			results[i] = true

		}
	}

	retvalue := true
	// We check all values. We are in a AND operator
	for i := 0; i < num_filters; i++ {
		retvalue = (retvalue && results[i])
	}

	return retvalue
}

// Update list of the UI depending VG list
func (MyUI *UI) UpdateList() {

	// Reset list of VG displayed on terminal
	MyUI.List.Rows = nil
	DisplayedVGPairing = nil

	// We select 1st line
	MyUI.List.SelectedRow = 0

	// We sort VG_List
	if MyUI.DisplayOrder == "name" {
		if MyUI.Order == "asc" {
			// asc on name
			sort.Slice(GlobalVGList, func(i, j int) bool {
				return GlobalVGList[i].Name < GlobalVGList[j].Name
			})
		} else {
			// desc on name
			sort.Slice(GlobalVGList, func(i, j int) bool {
				return GlobalVGList[i].Name > GlobalVGList[j].Name
			})
		}
	} else if MyUI.DisplayOrder == "uuid" {
		if MyUI.Order == "asc" {
			// asc on uuid
			sort.Slice(GlobalVGList, func(i, j int) bool {
				return GlobalVGList[i].UUID < GlobalVGList[j].UUID
			})
		} else {
			// desc on uuid
			sort.Slice(GlobalVGList, func(i, j int) bool {
				return GlobalVGList[i].UUID > GlobalVGList[j].UUID
			})
		}
	} else {
		panic("Unknown sort mode for VG list :" + MyUI.Order)
	}

	for i := 0; i < len(GlobalVGList); i++ {

		// We chack all VG fields
		if MyUI.Filter == "" || MyUI.MatchFilters(GlobalVGList[i]) {

			// We display the line if filter is ok
			MyUI.List.Rows = append(MyUI.List.Rows, MyUI.DisplayLine(i))
			DisplayedVGPairing = append(DisplayedVGPairing, i)
		}
	}

	// Get Number of VG displayed
	MyUI.TotalVG = len(DisplayedVGPairing)

	// Update list title
	MyUI.UpdateListTitle()
}

// Select all
func (MyUI *UI) SelectAll() {
	i := 0

	for i = 0; i < len(MyUI.List.Rows); i++ {
		index := DisplayedVGPairing[i]

		VGSelected[GlobalVGList[index].UUID] = true
	}

	tmp := 0
	for _, val := range VGSelected {
		if val {
			tmp++
		}
	}
	MyUI.SelectedItems = tmp
	MyUI.Log(fmt.Sprintf("All %d items are now selected", MyUI.SelectedItems), "yellow", "clear")
	MyUI.UpdateList()
}

// Change display mode
func (MyUI *UI) ChangeDisplayMode() {

	if MyUI.DisplayOrder == "uuid" {
		MyUI.List.Title = "VG List - Name (UUID)"
		MyUI.DisplayOrder = "name"
		MyUI.Log("Display modified from UUID (Name) to Name (UUID)", "yellow", "clear")
	} else if MyUI.DisplayOrder == "name" {
		MyUI.List.Title = "VG List - UUID (Name)"
		MyUI.Log("Display modified from Name (UUID) to UUID (Name)", "yellow", "clear")
		MyUI.DisplayOrder = "uuid"
	} else {
		panic("Display Order not handled")
	}
	MyUI.UpdateList()
}

// Function adding highlight for filters
func (MyUI *UI) AddFilterHighlight(Display string, filter [ConstMaxANDFiltering]*regexp.Regexp) string {
	tmp := ""

	for i := 0; i < ConstMaxANDFiltering; i++ {
		tmp = fmt.Sprintf("%s|%s|%s", filter[0].String(), filter[1].String(), filter[2].String())
	}

	tmpregexp, _ := regexp.Compile(tmp)

	return tmpregexp.ReplaceAllString(Display, "[$0]"+ConstFilterHighlight)
}

// Update detail part of the UI depending VG highlighted
func (MyUI UI) UpdateDetail() {

	// Now we display details if row exists
	if MyUI.List.Rows != nil {
		// We get the index of VG_List, with help of pairing array
		index := DisplayedVGPairing[MyUI.List.SelectedRow]

		// Now we display de detail content Line by line
		MyUI.Detail.Text = ""
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Cluster", MyUI.AddFilterHighlight(GlobalVGList[index].Cluster, MyUI.AdvFilter.Cluster))
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Name", MyUI.AddFilterHighlight(GlobalVGList[index].Name, MyUI.AdvFilter.Name))
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Container", MyUI.AddFilterHighlight(GlobalVGList[index].Container, MyUI.AdvFilter.Container))
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Size", MyUI.AddFilterHighlight(GlobalVGList[index].Size, MyUI.AdvFilter.Size))
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "UUID", MyUI.AddFilterHighlight(GlobalVGList[index].UUID, MyUI.AdvFilter.UUID))
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Mounted", MyUI.AddFilterHighlight(GlobalVGList[index].Attached, MyUI.AdvFilter.Mounted))
		MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Description", MyUI.AddFilterHighlight(GlobalVGList[index].Description, MyUI.AdvFilter.Description))
		
		// Add categories
		if len(GlobalVGList[index].Categories) > 0 {
			categoriesStr := categoriesMatchString(GlobalVGList[index].Categories)
			MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Categories", MyUI.AddFilterHighlight(categoriesStr, MyUI.AdvFilter.Categories))
		} else {
			MyUI.Detail.Text = fmt.Sprintf("%s%-12s: %s\n", MyUI.Detail.Text, "Categories", "-")
		}
	} else {
		MyUI.Detail.Text = "No row selected"
	}
}

// Update detail part of the UI depending VG highlighted
func (MyUI *UI) Select() {
	if MyUI.List.SelectedRow >= 0 {
		index := DisplayedVGPairing[MyUI.List.SelectedRow]

		if VGSelected[GlobalVGList[index].UUID] {
			VGSelected[GlobalVGList[index].UUID] = false
			MyUI.SelectedItems--
		} else {
			VGSelected[GlobalVGList[index].UUID] = true
			MyUI.SelectedItems++
		}
		MyUI.Log(fmt.Sprintf("Selected items : %d", MyUI.SelectedItems), "yellow", "clear")
		MyUI.List.Rows[MyUI.List.SelectedRow] = MyUI.DisplayLine(index)
	}
}

// Clear Selection
func (MyUI *UI) ClearSelection() {
	VGSelected = make(map[string]bool)
	MyUI.SelectedItems = 0
	MyUI.Log("VG selection cleared", "yellow", "clear")
	MyUI.UpdateList()
}

// Update list title
func (MyUI *UI) UpdateListTitle() {
	var txt string

	if MyUI.DisplayOrder == "name" {
		txt = "Name (UUID)"
	} else {
		txt = "UUID (Name)"
	}

	// Temporary variables for display
	MyUI.UpdateDetail()

	iconorder := ""
	if MyUI.Order == "asc" {
		iconorder = "\U00002B06"

	} else {
		iconorder = "\U00002B07"
	}

	// Update title
	MyUI.List.Title = fmt.Sprintf("VG List - %s - Total : %d/%d - Sorted : %s", txt, MyUI.TotalVG, len(GlobalVGList), iconorder)
}

// Order List
func (MyUI *UI) OrderList() {
	if MyUI.Order == "asc" {
		MyUI.Order = "desc"
		MyUI.Log("Display order change to descending", "yellow", "clear")
	} else {
		MyUI.Order = "asc"
		MyUI.Log("Display order change to ascending", "yellow", "clear")
	}

	// Update display
	MyUI.UpdateList()
}

func (MyUI *UI) PutFilterInAllFields(filter string, i int) {

	//Escape parenthesis
	var clean_filter string

	clean_filter = strings.Replace(filter, ")", "\\)", -1)
	clean_filter = strings.Replace(clean_filter, "(", "\\(", -1)

	MyUI.AdvFilter.UUID[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Name[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Cluster[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Mounted[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Description[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Size[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Container[i], _ = regexp.Compile(clean_filter)
	MyUI.AdvFilter.Categories[i], _ = regexp.Compile(clean_filter)
}

// Update Filterzone content
func (MyUI *UI) UpdateContentFilterZone(value string) {

	var clean_filter string

	// initialize backgrounf color
	background := ConstZoneActive
	error := false
	nb_filters := 0

	// We setup global variable
	MyUI.Filter = value

	// Clear filtering
	for i := 0; i < ConstMaxANDFiltering; i++ {
		MyUI.PutFilterInAllFields("###", i)
	}

	// update adv_filtering part of UI object
	list_filter := strings.Split(MyUI.Filter, "&")
	if len(list_filter) > ConstMaxANDFiltering {
		nb_filters = ConstMaxANDFiltering
		background = ConstZoneError
	} else {
		nb_filters = len(list_filter)
	}

	for i := 0; i < nb_filters; i++ {
		tmp := strings.Split(list_filter[i], ":")
		if len(tmp) == 1 {
			// Only a simple string :
			MyUI.PutFilterInAllFields(tmp[0], i)
		} else {
			clean_filter = strings.Replace(strings.Join(tmp[1:], ":"), ")", "\\)", -1)
			clean_filter = strings.Replace(clean_filter, "(", "\\(", -1)

			switch strings.ToLower(tmp[0]) {

			case "container":
				MyUI.AdvFilter.Container[i], _ = regexp.Compile(clean_filter)
			case "uuid":
				MyUI.AdvFilter.UUID[i], _ = regexp.Compile(clean_filter)
			case "name":
				MyUI.AdvFilter.Name[i], _ = regexp.Compile(clean_filter)
			case "cluster":
				MyUI.AdvFilter.Cluster[i], _ = regexp.Compile(clean_filter)
			case "mounted":
				MyUI.AdvFilter.Mounted[i], _ = regexp.Compile(clean_filter)
			case "description", "desc":
				MyUI.AdvFilter.Description[i], _ = regexp.Compile(clean_filter)
			case "size":
				MyUI.AdvFilter.Size[i], _ = regexp.Compile(clean_filter)
			case "categories", "category", "cat":
				MyUI.AdvFilter.Categories[i], _ = regexp.Compile(clean_filter)
			default:
				// Not a filed name, considered as simple string
				MyUI.PutFilterInAllFields(list_filter[i], i)
			}
		}
	}

	// Update filter zone in the UI
	if MyUI.Mode == "filter" {
		MyUI.FilterZone.Text = fmt.Sprintf("Filter : [%-30s]%s", MyUI.Filter, background)
	} else {
		if error {
			MyUI.FilterZone.Text = fmt.Sprintf("Filter : [%-30s]%s", MyUI.Filter, background)
		} else {
			MyUI.FilterZone.Text = fmt.Sprintf("Filter : [%-30s]%s", MyUI.Filter, ConstZonePassive)
		}
	}

	// Update Details zone
	MyUI.UpdateList()
	MyUI.UpdateDetail()
}

// Function displaying popup in the center of the screen
func (MyUI *UI) SetPopupSize(title string, x int, y int) {
	popup_width := x
	popup_height := y
	pos_x := (MyUI.TermWidth - popup_width) / 2
	pos_y := (MyUI.TermHeight - popup_height) / 2
	MyUI.Popup.Title = title
	MyUI.Popup.SetRect(pos_x, pos_y, pos_x+popup_width, pos_y+popup_height)
}

// Display Help Popup, if action is true : display, if false : hide
func (MyUI *UI) DisplayPopup(content string, action bool) {

	// Depending content
	switch content {
	case "help":
		// Action = True, I display the popup
		if action {
			MyUI.SetPopupSize("Help", 80, 27)

			MyUI.Popup.Text = "<Arrows>       : Move\n\nCtrl + A       : Select all items\n<Space> or S   : Select\nCtrl + <Space> : Clear selection\n\nF or /         : Filter\n                 You can specify fied (ex: Container:)\n                 | (or) and & (and) allowed\nCtrl + F       : Clear Filter\n\nD              : Change Display (uuid/name)\n\nO              : Change sort order\n\nU              : Update description\n\nCrtl + D       : Delete VG\n\nCrtl + R       : Refresh list\n\n\nEsc : Quit help"
			MyUI.Mode = "help"

		} else {
			MyUI.Popup.SetRect(0, 0, 0, 0)
			MyUI.Popup.Text = ""
			MyUI.Mode = "view"
		}
	}
	MyUI.Render()
}

// Display log message in the InteractBar
func (MyUI *UI) Log(content string, fgcolor string, bgcolor string) {

	tmp := ""

	// If term width is longer than content, we add
	if len(content) < (MyUI.TermWidth - 4) {
		tmp = content + strings.Repeat(" ", MyUI.TermWidth-len(content)-4)
	} else {
		tmp = content
	}

	MyUI.Interact.Text = fmt.Sprintf("[%s](fg:%s,bg:%s)", tmp, fgcolor, bgcolor)
	MyUI.Render()
}

// Display Asking Value
func (MyUI *UI) DisplayAskContent(text string, current_value string, fgcolor string, bgcolor string) {
	tmp := ""

	// If term width is longer than content, we add
	if len(current_value) < (MyUI.TermWidth - len(text) - 5) {
		tmp = current_value + strings.Repeat(" ", MyUI.TermWidth-len(current_value)-len(text)-7)
	} else {
		tmp = current_value
	}

	// We update interact bar
	MyUI.Interact.Text = fmt.Sprintf("%s : [%s](fg:%s,bg:%s)", text, tmp, fgcolor, bgcolor)
	MyUI.UpdateDetail()
	MyUI.Render()
}

// Ask information in the InteractBar
func (MyUI *UI) Ask(text string, authorized_chars string, fgcolor string, bgcolor string) string {

	var asking_value string

	// Because we are going to recreate a PoolEvents trigger here,
	// We need to inform the other one to flush 1 character
	// Seems to be a termUI bug.

	MyUI.UglyFix = true
	asking_value = ""

	// Display line in InteractBar
	MyUI.DisplayAskContent(text, asking_value, fgcolor, bgcolor)

	// Define authorize chars for filter
	reg, _ := regexp.Compile(authorized_chars)

	// We launch dedicated PollEvents
	uiEvents := ui.PollEvents()

	// Now we handle every intputs
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {

			// Validate
			case "<Enter>":
				ReturnValue := asking_value
				MyUI.Interact.Text = ""
				return ReturnValue

			// Cancel
			case "<C-c>", "<Escape>":
				MyUI.Interact.Text = ""
				return ""

			// Backspace
			case "<Backspace>":
				if len(asking_value) >= 1 {
					asking_value = asking_value[:len(asking_value)-1]
				}

			// Other chars
			default:
				if reg.MatchString(e.ID) {
					if e.ID[0:1] != "<" {
						// Add new char to filter
						asking_value += e.ID
					} else if e.ID == "<Space>" {
						asking_value += " "
					}
				}
			}

			// We display value in the InteractBar
			MyUI.DisplayAskContent(text, asking_value, fgcolor, bgcolor)
		}

	}
}

// Ask information in the InteractBar
func (MyUI *UI) AskFilter(text string, authorized_chars string, fgcolor string, bgcolor string) string {

	var asking_value string

	// Because we are going to recreate a PoolEvents trigger here,
	// We need to inform the other one to flush 1 character
	// Seems to be a termUI bug.

	MyUI.UglyFix = true

	// Intialize asking_value with current filter
	asking_value = MyUI.Filter

	// Display line in InteractBar
	MyUI.DisplayAskContent(text, asking_value, fgcolor, bgcolor)

	// Define authorize chars for filter
	reg, _ := regexp.Compile(authorized_chars)

	// We launch dedicated PollEvents
	uiEvents := ui.PollEvents()

	// Now we handle every intputs
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {

			// Validate
			case "<Enter>":
				ReturnValue := asking_value
				MyUI.Interact.Text = ""
				return ReturnValue

			// Cancel
			case "<C-c>", "<Escape>":
				MyUI.Interact.Text = ""
				return ""

			// Backspace
			case "<Backspace>":
				if len(asking_value) >= 1 {
					asking_value = asking_value[:len(asking_value)-1]
				}

			// Tab
			case "<Tab>":
				// We extract last chars until & or |
				pos1 := strings.LastIndex(asking_value, "|")
				pos2 := strings.LastIndex(asking_value, "&")
				pos := 0

				if (pos1 == -1) && (pos2 == -1) {
					pos = -1
				} else if pos1 > pos2 {
					pos = pos1
				} else {
					pos = pos2
				}

				//Extract substring
				tmpsub := string(asking_value[pos+1:])
				tmpfield := ""

				// If filter contains beginning of VG struct field, we store the field name
				switch {
				case strings.HasPrefix("uuid", strings.ToLower(tmpsub)):
					tmpfield = "uuid"
				case strings.HasPrefix("container", strings.ToLower(tmpsub)):
					tmpfield = "container"
				case strings.HasPrefix("name", strings.ToLower(tmpsub)):
					tmpfield = "name"
				case strings.HasPrefix("cluster", strings.ToLower(tmpsub)):
					tmpfield = "cluster"
				case strings.HasPrefix("size", strings.ToLower(tmpsub)):
					tmpfield = "size"
				case strings.HasPrefix("description", strings.ToLower(tmpsub)):
					tmpfield = "description"
				case strings.HasPrefix("mounted", strings.ToLower(tmpsub)):
					tmpfield = "mounted"
				case strings.HasPrefix("categories", strings.ToLower(tmpsub)):
					tmpfield = "categories"
				}

				// If tmpfield exists, we complete the filter
				if tmpfield != "" {
					asking_value += tmpfield[len(tmpsub):] + ":"
				}

			// Other chars
			default:
				if reg.MatchString(e.ID) {
					if e.ID[0:1] != "<" {
						// Add new char to filter
						asking_value += e.ID
					} else if e.ID == "<Space>" {
						asking_value += " "
					}
				}
			}

			MyUI.UpdateContentFilterZone(asking_value)

			// We display value in the InteractBar
			MyUI.DisplayAskContent(text, asking_value, fgcolor, bgcolor)
		}

	}
}

// Update Description
func (MyUI *UI) UpdateVGDescription() {

	NewDesc := MyUI.Ask("New description to set", `[A-Za-z0-9\-\:\|\&\_\,\ ]`, "black", "green")
	Confirm := MyUI.Ask(fmt.Sprintf("Please write ['CONFIRM'](mod:bold) to update %d VG", MyUI.SelectedItems), `[A-Za-z0-9\-\:\|\&\_\,\ ]`, "black", "green")

	if Confirm == "CONFIRM" {

		MyUI.Log("Please wait during description update...", "yellow", "clear")

		for uuid, state := range VGSelected {
			if state {
				payload := "{\"description\":\"" + NewDesc + "\",\"extId\":\"" + uuid + "\"}"
				MyPrism.CallAPIJSON("PC", "PATCH", "/api/storage/v4.0.a2/config/volume-groups/"+uuid, payload, nil)
			}
		}

		// Finished
		MyUI.Log("Description update done... Launching VG list refresh...", "green", "clear")

		// update list
		GetVGList()
		MyUI.UpdateList()

	} else {
		MyUI.Log("Wrong confirmation word. Operation cancelled", "red", "clear")
	}
}

// Update Description
func (MyUI *UI) RequestDeleteVG() bool {

	Confirm := MyUI.Ask(fmt.Sprintf("Deletion of %d VG, data will not be retievable. Write ['I UNDERSTAND'](mod:bold) to confirm", MyUI.SelectedItems), `[A-Za-z0-9\-\:\|\&\_\,\ ]`, "black", "green")

	if strings.ToUpper(Confirm) == "I UNDERSTAND" {

		MyUI.Log("Please wait during VG deletion (can take a while)...", "yellow", "clear")

		count := 0

		for uuid, state := range VGSelected {
			if state {
				count++
				MyUI.Log(fmt.Sprintf("Deletion VG %d/%d, please wait...", count, MyUI.SelectedItems), "yellow", "clear")
				status := DeleteVG(uuid)

				if status {
					MyUI.Log("VG "+uuid+" succesfuly deleted...", "green", "clear")
				} else {
					MyUI.Log("Unable to delete VG "+uuid, "red", "clear")
				}
			}
		}

		// Finished
		MyUI.Log("Deletion done... Soft-updating list...", "green", "clear")

		MyUI.UpdateList()
		MyUI.Log("List updated...", "green", "clear")
		return true

	} else {
		MyUI.Log("Incorrect confirmation : operation canceled.", "red", "clear")
		return false
	}
}

// EnterFilter
func (MyUI *UI) EnterFilter() {
	tmp := MyUI.AskFilter("Enter your filter", `^[A-Za-z0-9\-\:\|\&\,\(\)]$|<Space>`, "black", "green")
	MyUI.Filter = tmp
}