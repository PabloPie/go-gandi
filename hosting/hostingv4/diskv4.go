package hostingv4

import (
	"strconv"

	"github.com/PabloPie/Gandi-Go/hosting"
)

type (
	DiskSpec   = hosting.DiskSpec
	Disk       = hosting.Disk
	DiskImage  = hosting.DiskImage
	DiskFilter = hosting.DiskFilter
)

type diskv4 struct {
	ID       int    `xmlrpc:"id"`
	Name     string `xmlrpc:"name"`
	Size     int    `xmlrpc:"size"`
	RegionID int    `xmlrpc:"datacenter_id"`
	State    string `xmlrpc:"state"`
	Type     string `xmlrpc:"type"`
	VM       []int  `xmlrpc:"vms_id"`
	BootDisk bool   `xmlrpc:"is_boot_disk"`
}

type diskSpecv4 struct {
	RegionID int    `xmlrpc:"datacenter_id"`
	Name     string `xmlrpc:"name"`
	Size     int    `xmlrpc:"size"`
}

type diskFilterv4 struct {
	ID       int    `xmlrpc:"id"`
	RegionID int    `xmlrpc:"datacenter_id"`
	Name     string `xmlrpc:"name"`
	VMID     int    `xmlrpc:"vm_id"`
}

// CreateDisk creates a new empty disk.
// If left unspecified, newDisks's `Name` will be generated by Gandi's API,
// `Size` will default to 10GB and `Region` to FR_SD3
func (h Hostingv4) CreateDisk(newDisk DiskSpec) (Disk, error) {
	diskv4, err := toDiskSpecv4(newDisk)
	if err != nil {
		return Disk{}, err
	}
	disk, _ := structToMap(diskv4)

	response := Operation{}
	params := []interface{}{disk}
	err = h.Send("hosting.disk.create", params, &response)
	if err != nil {
		return Disk{}, err
	}
	if err := h.waitForOp(response); err != nil {
		return Disk{}, err
	}

	return h.diskFromID(response.DiskID)
}

// CreateDiskFromImage creates a disk with the same data as `srcDisk`. If `Size` is
// not specified for `newDisk`, it will be created with size `srcDisk.Size`
func (h Hostingv4) CreateDiskFromImage(newDisk DiskSpec, srcDisk DiskImage) (Disk, error) {
	var fn = "CreateDiskFromImage"
	if srcDisk.DiskID == "" {
		return Disk{}, &HostingError{fn, "DiskImage", "DiskID", ErrNotProvided}
	}

	diskv4, err := toDiskSpecv4(newDisk)
	if err != nil {
		return Disk{}, err
	}
	disk, _ := structToMap(diskv4)
	imageid, err := strconv.Atoi(srcDisk.DiskID)
	if err != nil {
		return Disk{}, &HostingError{fn, "DiskImage", "DiskID", ErrParse}
	}

	response := Operation{}
	params := []interface{}{disk, imageid}
	err = h.Send("hosting.disk.create_from", params, &response)
	if err != nil {
		return Disk{}, err
	}
	if err := h.waitForOp(response); err != nil {
		return Disk{}, err
	}

	return h.diskFromID(response.DiskID)
}

// ListDisks lists every disk
func (h Hostingv4) ListDisks() ([]Disk, error) {
	return h.DescribeDisks(DiskFilter{})
}

// DiskFromID is a helper function to get a Disk given its ID,
// if the vm does not exist or an error ocurred it returns an empty Disk,
// use DescribeDisks with an appropriate DiskFilter to get more details
// on the possible errors
func (h Hostingv4) DiskFromID(id string) Disk {
	disks, err := h.DescribeDisks(DiskFilter{ID: id})
	if err != nil || len(disks) < 1 {
		return Disk{}
	}

	return disks[0]
}

// DescribeDisks return a list of disks filtered with the options provided in `diskFilter`
func (h Hostingv4) DescribeDisks(diskfilter DiskFilter) ([]Disk, error) {
	filterv4, err := toDiskFilterv4(diskfilter)
	if err != nil {
		return nil, err
	}
	filter, _ := structToMap(filterv4)

	response := []diskv4{}
	params := []interface{}{}
	if len(filter) > 0 {
		params = append(params, filter)
	}
	// disk.list and disk.info return the same information
	err = h.Send("hosting.disk.list", params, &response)
	if err != nil {
		return nil, err
	}

	var disks []Disk
	for _, disk := range response {
		disks = append(disks, fromDiskv4(disk))
	}
	return disks, nil
}

// DeleteDisk deletes the Disk `disk`
func (h Hostingv4) DeleteDisk(disk Disk) error {
	var fn = "DeleteDisk"
	if disk.ID == "" {
		return &HostingError{fn, "Disk", "ID", ErrNotProvided}
	}

	diskid, err := strconv.Atoi(disk.ID)
	if err != nil {
		return &HostingError{fn, "Disk", "ID", ErrParse}
	}

	response := Operation{}
	params := []interface{}{diskid}
	err = h.Send("hosting.disk.delete", params, &response)
	if err != nil {
		return err
	}
	err = h.waitForOp(response)
	return err
}

// ExtendDisk extends `disk.Size` by `size` (original size + `size`),
// Disks cannot shrink in size, `size` is in GB
func (h Hostingv4) ExtendDisk(disk Disk, size uint) (Disk, error) {
	var fn = "ExtendDisk"
	if disk.ID == "" {
		return Disk{}, &HostingError{fn, "Disk", "ID", ErrNotProvided}
	}
	// size has to be a multiple of 1024
	newSize := disk.Size + size*1024
	diskupdate := map[string]int{"size": int(newSize)}
	diskid, err := strconv.Atoi(disk.ID)
	if err != nil {
		return Disk{}, &HostingError{fn, "Disk", "ID", ErrParse}
	}

	response := Operation{}
	request := []interface{}{diskid, diskupdate}
	err = h.Send("hosting.disk.update", request, &response)
	if err != nil {
		return Disk{}, err
	}
	if err := h.waitForOp(response); err != nil {
		return Disk{}, err
	}

	return h.diskFromID(response.DiskID)
}

// RenameDisk changes the name of `disk` to `newName`
func (h Hostingv4) RenameDisk(disk Disk, newName string) (Disk, error) {
	var fn = "RenameDisk"
	if disk.ID == "" {
		return Disk{}, &HostingError{fn, "Disk", "ID", ErrNotProvided}
	}
	diskupdate := map[string]string{"name": newName}
	diskid, err := strconv.Atoi(disk.ID)
	if err != nil {
		return Disk{}, &HostingError{fn, "Disk", "ID", ErrParse}
	}

	response := Operation{}
	request := []interface{}{diskid, diskupdate}
	err = h.Send("hosting.disk.update", request, &response)
	if err != nil {
		return Disk{}, err
	}
	err = h.waitForOp(response)
	if err != nil {
		return Disk{}, err
	}

	return h.diskFromID(response.DiskID)
}

// Helper functions

// Obtain a Hosting Disk from an integer ID (v4 representation)
func (h Hostingv4) diskFromID(id int) (Disk, error) {
	response := diskv4{}
	params := []interface{}{id}
	err := h.Send("hosting.disk.info", params, &response)
	if err != nil {
		return Disk{}, err
	}
	disk := fromDiskv4(response)
	return disk, nil
}

// Conversion functions for Disks in Gandi v4

// toDiskFilterv4 converts a Hosting DiskSpec to a v4 DiskSpec
func toDiskSpecv4(disk DiskSpec) (diskSpecv4, error) {
	region, err := strconv.Atoi(disk.RegionID)
	if err != nil {
		return diskSpecv4{}, internalParseError("DiskSpec", "RegionID")
	}
	return diskSpecv4{
		RegionID: region,
		Name:     disk.Name,
		Size:     int(disk.Size),
	}, nil
}

// isIgnorableErr return true if an Atoi conversion error is caused
// by a bad string format (ErrSyntax) or there is no error
func isIgnorableErr(err error) bool {
	if err != nil {
		err = err.(*strconv.NumError).Err
		return err == strconv.ErrSyntax
	}
	return true
}

// toDiskFilterv4 converts a Hosting DiskFilter to a v4 DiskFilter
func toDiskFilterv4(disk DiskFilter) (diskFilterv4, error) {
	region, err := strconv.Atoi(disk.RegionID)
	if !isIgnorableErr(err) {
		return diskFilterv4{}, internalParseError("DiskFilter", "RegionID")
	}

	id, err := strconv.Atoi(disk.ID)
	if !isIgnorableErr(err) {
		return diskFilterv4{}, internalParseError("DiskFilter", "ID")
	}

	vmid, err := strconv.Atoi(disk.VMID)
	if !isIgnorableErr(err) {
		return diskFilterv4{}, internalParseError("DiskFilter", "VMID")
	}
	return diskFilterv4{
		RegionID: region,
		ID:       id,
		VMID:     vmid,
	}, nil
}

// fromDiskv4 transforms a v4 Disk to a Hosting Disk, casting integers to strings
func fromDiskv4(disk diskv4) Disk {
	id := strconv.Itoa(disk.ID)
	region := strconv.Itoa(disk.RegionID)
	var vms []string
	for _, vm := range disk.VM {
		vm := strconv.Itoa(vm)
		vms = append(vms, vm)
	}
	return Disk{
		ID:       id,
		Name:     disk.Name,
		Size:     uint(disk.Size),
		RegionID: region,
		State:    disk.State,
		Type:     disk.Type,
		VM:       vms,
		BootDisk: disk.BootDisk,
	}
}