package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Declare VG struct
type VG struct {
	Cluster        string
	UUID           string
	Name           string
	Container      string
	Description    string
	Size           string
	Attached       string
	Attached_vm    []string
	Attached_iscsi []string
}

// Create global GlobalVGList
var GlobalVGList []VG

// function to get list (without details, we want quick operation)
func GetVGList() {
	var StoContainerList = make(map[string]string) // UUID => Storage Container name
	var VM_List = make(map[string]string)          // UUID => VM Name
	var iSCSI_List = make(map[string]string)       // UUID => iSCSI

	// Reinit VG_List
	GlobalVGList = nil

	// =========================== Get Storage Container List ===========================
	var tmp1 struct {
		Data []struct {
			ContainerExtID string `json:"containerExtId"`
			Name           string `json:"name"`
		} `json:"data"`
	}

	MyPrism.CallAPIJSON("PC", "GET", "/api/storage/v4.0.a2/config/storage-containers", "", &tmp1)

	// Parse all SC
	for tmp := range tmp1.Data {
		StoContainerList[tmp1.Data[tmp].ContainerExtID] = tmp1.Data[tmp].Name
	}

	// =========================== Get external iSCSI List ===========================
	var tmpiscsi struct {
		DataItemDiscriminator string `json:"$dataItemDiscriminator"`
		Data                  []struct {
			ExtID              string `json:"extId"`
			IscsiInitiatorName string `json:"iscsiInitiatorName"`
		} `json:"data"`
	}

	// We do request until we get EMPTY_LIST
	for page := 0; tmpiscsi.DataItemDiscriminator != "EMPTY_LIST"; page++ {

		MyPrism.CallAPIJSON("PC", "GET", fmt.Sprintf("/api/storage/v4.0.a2/config/iscsi-clients?$limit=100&$page=%d", page), "", &tmpiscsi)
		if tmpiscsi.DataItemDiscriminator != "EMPTY_LIST" {
			// If list is not empty, we store everything in an array
			for _, tmp := range tmpiscsi.Data {
				iSCSI_List[tmp.ExtID] = tmp.IscsiInitiatorName
			}
		}
	}

	// =========================== Get VM list ===========================
	var tmpvm struct {
		GroupResults []struct {
			EntityResults []struct {
				Data []struct {
					DataType string `json:"data_type"`
					Name     string `json:"name"`
					Values   []struct {
						Time   int64    `json:"time"`
						Values []string `json:"values"`
					} `json:"values"`
				} `json:"data"`
				EntityID string `json:"entity_id"`
			} `json:"entity_results"`
		} `json:"group_results"`
	}

	payloadvm := `{
		"entity_type":"mh_vm",
		"group_member_attributes":[
		   {
			  "attribute":"vm_name"
		   }
		]
	 }`

	MyPrism.CallAPIJSON("PC", "POST", "/api/nutanix/v3/groups", payloadvm, &tmpvm)

	for _, tmp := range tmpvm.GroupResults[0].EntityResults {

		vmname := ""

		for _, tmp2 := range tmp.Data {
			switch tmp2.Name {
			case "vm_name":
				vmname = tmp2.Values[0].Values[0]
			}

			VM_List[tmp.EntityID] = vmname

		}
	}

	// =========================== Get VG list ===========================
	var VGList struct {
		GroupResults []struct {
			EntityResults []struct {
				Data []struct {
					DataType string `json:"data_type"`
					Name     string `json:"name"`
					Values   []struct {
						Time   int64    `json:"time"`
						Values []string `json:"values"`
					} `json:"values"`
				} `json:"data"`
				EntityID string `json:"entity_id"`
			} `json:"entity_results"`
		} `json:"group_results"`
	}

	// We do the API call to get VG list
	payload := `{
        "entity_type": "volume_group_config",
        "group_member_attributes": [
            {
                "attribute": "name"
            },
            {
                "attribute": "controller_user_bytes"
            },
            {
                "attribute": "client_uuids"
            },
            {
                "attribute": "cluster_name"
            },
            {
                "attribute": "capacity_bytes"
            },
            {
                "attribute": "vm_uuids"
            },
                        {
                "attribute": "annotation"
            },
                                    {
                "attribute": "container_uuids"
            }
        ]
    }`
	MyPrism.CallAPIJSON("PC", "POST", "/api/nutanix/v3/groups", payload, &VGList)

	// Parse all VG an put them in a file
	for _, tmp := range VGList.GroupResults[0].EntityResults {

		// Create vg
		var tmpelt VG
		var sto_used, sto_capacity float64

		for _, tmp2 := range tmp.Data {

			// Fill struct elements regarding "name" field
			switch tmp2.Name {
			case "name":
				if len(tmp2.Values) > 0 {
					tmpelt.Name = tmp2.Values[0].Values[0]
				}
			case "controller_user_bytes":
				if len(tmp2.Values) > 0 {

					conv, _ := strconv.Atoi(tmp2.Values[0].Values[0])
					sto_used = float64(conv) / (1024 * 1024 * 1024)
				}
			case "client_uuids":
				if len(tmp2.Values) > 0 {

					tmpelt.Attached_iscsi = tmp2.Values[0].Values
				}
			case "cluster_name":
				if len(tmp2.Values) > 0 {

					tmpelt.Cluster = tmp2.Values[0].Values[0]
				}
			case "capacity_bytes":
				if len(tmp2.Values) > 0 {

					conv, _ := strconv.Atoi(tmp2.Values[0].Values[0])
					sto_capacity = float64(conv) / (1024 * 1024 * 1024)
				}
			case "vm_uuids":
				if len(tmp2.Values) > 0 {

					tmpelt.Attached_vm = tmp2.Values[0].Values
				}
			case "annotation":
				if len(tmp2.Values) > 0 {

					tmpelt.Description = tmp2.Values[0].Values[0]
				}
			case "container_uuids":
				if len(tmp2.Values) > 0 {

					tmpelt.Container = StoContainerList[tmp2.Values[0].Values[0]]
				}
			}

		}
		tmpelt.Size = fmt.Sprintf("%0.2f (%0.2f used)", sto_capacity, sto_used)

		if len(tmpelt.Attached_vm) > 0 || len(tmpelt.Attached_iscsi) > 0 {
			var tmp []string
			var j_vm string = ""
			var j_iscsi string = ""

			if len(tmpelt.Attached_vm) > 0 {
				tmp = []string{}
				for _, vmuuid := range tmpelt.Attached_vm {
					tmp = append(tmp, VM_List[vmuuid])
				}
				j_vm = strings.Join(tmp, ",")
			} else {
				j_vm = "-"
			}

			if len(tmpelt.Attached_iscsi) > 0 {
				tmp = []string{}
				for _, iscsi_uuid := range tmpelt.Attached_iscsi {
					tmp = append(tmp, iSCSI_List[iscsi_uuid])
				}
				j_iscsi = strings.Join(tmp, ",")
			} else {
				j_iscsi = "-"
			}

			tmpelt.Attached = fmt.Sprintf("True (VM : %s / iSCSI : %s)", j_vm, j_iscsi)
		} else {
			tmpelt.Attached = "False"
		}
		tmpelt.UUID = tmp.EntityID

		GlobalVGList = append(GlobalVGList, tmpelt)
	}

}

// DeleteVG UUID
func DeleteVG(uuid string) bool {
	var vg VG
	var IndexOfVG int

	// We start by finding this UUID in VG_List
	for IndexOfVG, vg = range GlobalVGList {
		if vg.UUID == uuid {
			break // for loop
		}
	}

	var answer struct {
		Data struct {
			ExtID string `json:"extId"`
		} `json:"data"`
	}

	// We start by detaching iscsi connection if exists
	for _, tmp := range vg.Attached_iscsi {
		MyPrism.CallAPIJSON("PC", "POST", "/api/storage/v4.0.a2/config/volume-groups/"+vg.UUID+"/$actions/detach-iscsi-client/"+tmp, "", &answer)

		tmp2 := strings.Split(answer.Data.ExtID, ":")
		state := WaitForTask(tmp2[1])

		if !state {
			return false
		}
	}

	// We continue by detaching vm connection if exists
	for _, tmp := range vg.Attached_vm {
		MyPrism.CallAPIJSON("PC", "POST", "/api/storage/v4.0.a2/config/volume-groups/"+vg.UUID+"/$actions/detach-vm/"+tmp, "", &answer)

		tmp2 := strings.Split(answer.Data.ExtID, ":")
		state := WaitForTask(tmp2[1])

		if !state {
			return false
		}
	}

	// Now we can delete the VG
	MyPrism.CallAPIJSON("PC", "DELETE", "/api/storage/v4.0.a2/config/volume-groups/"+vg.UUID, "", &answer)

	tmp2 := strings.Split(answer.Data.ExtID, ":")
	state := WaitForTask(tmp2[1])

	// Now we can delete VG from list
	SmartDelete(IndexOfVG)

	if !state {
		// Deletion failed
		return false
	} else {
		// Deletion OK
		return true
	}
}

// Function WaitForTask
func WaitForTask(taskID string) bool {

	var answer struct {
		Status             string `json:"status"`
		PercentageComplete int    `json:"percentage_complete"`
	}

	MyPrism.CallAPIJSON("PC", "GET", "/api/nutanix/v3/tasks/"+taskID, "", &answer)

	// We wait for end of task
	for answer.PercentageComplete < 100 {
		time.Sleep(time.Duration(3) * time.Second)
		MyPrism.CallAPIJSON("PC", "GET", "/api/nutanix/v3/tasks/"+taskID, "", &answer)
	}

	if answer.Status == "SUCCEEDED" {
		return true
	} else {
		return false
	}
}

// SmartDelete
// Delete single entry from VGList without reload all VG
func SmartDelete(index int) {
	GlobalVGList = append(GlobalVGList[:index], GlobalVGList[index+1:]...)

}
