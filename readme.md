# Nutanix VG Manager (NVGM)

## Summary ##

nvgm is a tool to manage VG from Nutanix clusters. It allows to list and display and filter them, display all details of the VG, change their description or delete VG, one by one or in bulk mode.

This tool is NOT an official tool of Nutanix.

## Demo


## Install

### Brew
With brew, just do 
```
brew install golgautier/tap/nvgm
```

### Brew
Download your binary corresponding to your platform from the active release

### Compile your own binary ###
If you prefer to create your own binary (for security reason). Download code from this folder, install golang on your computer, and launch command

`go build -o <binary name> .`


## Launch ##

Interactive : 
just launch `nvgm` and answer to all questions.

With parameters :
Launch `nvgm` with several parameters :
- `--help` or `-usage`: Display usage message
- `--secure-mode`: Secure mode, to force https with valid certificate
- `--server <Server name>` : Specify Prism Central name or address
- `--user <User name>` : Specify user to use (must have admin rights) 
- `--debug-mode` : create debug.log fiel with all API calls done to PC
- `--old-pc` : if you have PC before 2024.1, please use this flag

Password will be requested interactively

Example : 
```
./nvgm -server pc.ntnx.fr --user admin
```

## Usage ##

When app is launched, it will get VG list from PC. It can take few seconds regarding VG numbers

Then, VG are displayed in an array. You can use the following keys :
- `h` : Display help popup (then `Escape` to leave it)
- `d` : Switch display mode : Name (UUID) or UUID (Name)
- `o` : Order list ascending or descending 
- `f` : filter VG List (Warning : it is case sentsitive)
  - You can use simple expression, checked on all VG details
  - You can specify field to filter (`UUID`, `Container`, `Name`, `Size`, `Mounted`, `Description`) by adding field name then `:` and the filter value
     - `Tab` do auto-completion, for the field name
  - You can specify multiple filter values with `|` (OR) or `&` (AND)
- `Space` : Select the highlighted VG
- `Ctrl + A` : Select all displayed VG
- `Ctrl + Space` : Clear Selection
- `u` : Update description of selected VG (or the highlighted one)
- `Ctrl + d` : Delete selected VG (or the highlighted one)
- `q` : Quit

