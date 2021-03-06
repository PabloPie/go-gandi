package hostingv4

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/PabloPie/go-gandi/client"
	"github.com/PabloPie/go-gandi/hosting"
	"github.com/PabloPie/go-gandi/mock"
	"github.com/golang/mock/gomock"
)

var (
	// default values
	defaultRegion = 3
	defaultSize   = 10
	defaultSizeMB = 10240
	// expected params
	diskid     = 1
	diskidstr  = "1"
	imageid    = 100
	imageidstr = "100"
	diskname   = "Disk1"
	disksize   = 20
	disksizeMB = 20480
	region     = 4
	regionstr  = "4"
)

var disks = []diskv4{
	{1, "sys_disk1", 10240, 4, "created", "data", []int{1}, true},
	{4, "sys_disk3", 10240, 4, "created", "data", []int{3}, true},
	{2, "sys_disk2", 20480, 3, "created", "data", []int{2}, true},
	{3, "disk3", 10240, 3, "created", "data", []int{2}, false},
	{5, diskname, 10240, 3, "created", "data", []int{}, false},
}

func TestCreateDiskWithNameSizeAndRegion(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskCreate := []interface{}{map[string]interface{}{
		"datacenter_id": region,
		"name":          diskname,
		"size":          disksizeMB,
	}}
	responseDiskCreate := Operation{
		ID:     1,
		DiskID: diskid,
	}
	creation := mockClient.EXPECT().Send("hosting.disk.create",
		paramsDiskCreate, gomock.Any()).SetArg(2, responseDiskCreate).Return(nil)

	paramsWait := []interface{}{responseDiskCreate.ID}
	responseWait := operationInfo{responseDiskCreate.ID, "DONE"}
	wait := mockClient.EXPECT().Send("operation.info",
		paramsWait, gomock.Any()).SetArg(2, responseWait).Return(nil).After(creation)

	paramsDiskInfo := []interface{}{responseDiskCreate.DiskID}
	responseDiskInfo := diskv4{diskid, diskname, disksizeMB, region, "created", "data", []int{}, false}
	mockClient.EXPECT().Send("hosting.disk.info",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil).After(wait)

	diskspec := hosting.DiskSpec{
		RegionID: regionstr,
		Name:     diskname,
		Size:     disksize,
	}
	disk, _ := testHosting.CreateDisk(diskspec)

	expected := hosting.Disk{
		ID:       diskidstr,
		Name:     diskname,
		Size:     disksize,
		RegionID: regionstr,
		State:    "created",
		Type:     "data",
		BootDisk: false,
	}

	if !reflect.DeepEqual(disk, expected) {
		t.Errorf("Error, expected %+v, got instead %+v", expected, disk)
	}
}

func TestCreateDiskFromImageWithoutSize(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskCreate := []interface{}{map[string]interface{}{
		"datacenter_id": region,
		"name":          diskname,
	}, imageid}
	responseDiskCreate := Operation{
		ID:     1,
		DiskID: diskid,
	}
	creation := mockClient.EXPECT().Send("hosting.disk.create_from",
		paramsDiskCreate, gomock.Any()).SetArg(2, responseDiskCreate).Return(nil)

	paramsWait := []interface{}{responseDiskCreate.ID}
	responseWait := operationInfo{responseDiskCreate.ID, "DONE"}
	wait := mockClient.EXPECT().Send("operation.info",
		paramsWait, gomock.Any()).SetArg(2, responseWait).Return(nil).After(creation)

	paramsDiskInfo := []interface{}{responseDiskCreate.DiskID}
	responseDiskInfo := diskv4{diskid, diskname, 3072, region, "created", "data", []int{}, false}
	mockClient.EXPECT().Send("hosting.disk.info",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil).After(wait)

	diskspec := hosting.DiskSpec{
		RegionID: regionstr,
		Name:     diskname,
	}
	diskimage := hosting.DiskImage{
		DiskID:   imageidstr,
		Size:     3,
		Name:     "Debian 9",
		RegionID: regionstr,
	}
	disk, _ := testHosting.CreateDiskFromImage(diskspec, diskimage)

	expected := hosting.Disk{
		ID:       diskidstr,
		Name:     diskname,
		Size:     3,
		RegionID: regionstr,
		State:    "created",
		Type:     "data",
		BootDisk: false,
	}

	if !reflect.DeepEqual(disk, expected) {
		t.Errorf("Error, expected %+v, got instead %+v", expected, disk)
	}
}

func TestListAllDisks(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskInfo := []interface{}{}
	responseDiskInfo := disks
	mockClient.EXPECT().Send("hosting.disk.list",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil)

	disks, _ := testHosting.ListAllDisks()

	if len(disks) < 1 {
		t.Errorf("Error, expected to get at least 1 hosting.Disk")
	}
}

func TestListDiskWithEmptyFilterNodisks(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskInfo := []interface{}{}
	responseDiskInfo := []diskv4{}
	mockClient.EXPECT().Send("hosting.disk.list",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil)

	diskfilter := hosting.DiskFilter{}
	disks, _ := testHosting.ListDisks(diskfilter)

	if len(disks) > 0 {
		t.Errorf("Error, expected to get no Disks")
	}
}

func TestListDiskWithNameInFilter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskInfo := []interface{}{map[string]interface{}{
		"name": diskname,
	}}
	responseDiskInfo := disks[4:]
	mockClient.EXPECT().Send("hosting.disk.list",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil)

	diskfilter := hosting.DiskFilter{Name: diskname}
	disks, _ := testHosting.ListDisks(diskfilter)

	if len(disks) != 1 {
		t.Errorf("Error, expected to get only 1 hosting.Disk and got %d instead", len(disks))
	}
	if disks[0].Name != diskname {
		t.Errorf("Error, expected to get hosting.Disk with name '%s', got '%s' instead",
			diskname, disks[0].Name)
	}
}

func TestDiskFromName(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskInfo := []interface{}{map[string]interface{}{
		"name": diskname,
	}}
	responseDiskInfo := disks[4:]
	mockClient.EXPECT().Send("hosting.disk.list",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil)

	disk := testHosting.DiskFromName(diskname)

	if disk.Name != diskname {
		t.Errorf("Error, expected to get hosting.Disk with name '%s', got '%s' instead",
			diskname, disks[0].Name)
	}
}

func TestDeleteDisk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	paramsDiskDelete := []interface{}{diskid}
	responseDiskDelete := Operation{
		ID:     1,
		DiskID: diskid,
	}
	delete := mockClient.EXPECT().Send("hosting.disk.delete",
		paramsDiskDelete, gomock.Any()).SetArg(2, responseDiskDelete).Return(nil)

	paramsWait := []interface{}{responseDiskDelete.ID}
	responseWait := operationInfo{responseDiskDelete.ID, "DONE"}
	mockClient.EXPECT().Send("operation.info",
		paramsWait, gomock.Any()).SetArg(2, responseWait).Return(nil).After(delete)

	err := testHosting.DeleteDisk(hosting.Disk{ID: diskidstr})
	if err != nil {
		t.Errorf("Error, expected disk to be deleted, got error '%v' instead", err)
	}
}
func TestExtendDisk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	var extendSize uint = 5
	var sizeInMB = 5120

	paramsDiskExtend := []interface{}{
		disks[0].ID,
		map[string]int{"size": disks[0].Size + sizeInMB},
	}
	responseDiskExtend := Operation{
		ID:     1,
		DiskID: disks[0].ID,
	}
	delete := mockClient.EXPECT().Send("hosting.disk.update",
		paramsDiskExtend, gomock.Any()).SetArg(2, responseDiskExtend).Return(nil)

	paramsWait := []interface{}{responseDiskExtend.ID}
	responseWait := operationInfo{responseDiskExtend.ID, "DONE"}
	wait := mockClient.EXPECT().Send("operation.info",
		paramsWait, gomock.Any()).SetArg(2, responseWait).Return(nil).After(delete)

	paramsDiskInfo := []interface{}{responseDiskExtend.DiskID}
	responseDiskInfo := disks[0]
	responseDiskInfo.Size += sizeInMB
	mockClient.EXPECT().Send("hosting.disk.info",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil).After(wait)

	diskparam := hosting.Disk{ID: strconv.Itoa(disks[0].ID), Size: 10}
	disk, _ := testHosting.ExtendDisk(diskparam, extendSize)
	expectedDiskSize := (disks[0].Size + sizeInMB) / 1024
	if disk.Size != expectedDiskSize {
		t.Errorf("Error, expected disk size %dGB, got a size of %dGB instead",
			disks[0].Size+sizeInMB, disk.Size)
	}
}

func TestRenameDisk(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mock.NewMockV4Caller(mockCtrl)
	testHosting := Newv4Hosting(mockClient)

	var newName = "Disk_renamed"

	paramsDiskExtend := []interface{}{
		disks[0].ID,
		map[string]string{"name": newName},
	}
	responseDiskExtend := Operation{
		ID:     1,
		DiskID: disks[0].ID,
	}
	delete := mockClient.EXPECT().Send("hosting.disk.update",
		paramsDiskExtend, gomock.Any()).SetArg(2, responseDiskExtend).Return(nil)

	paramsWait := []interface{}{responseDiskExtend.ID}
	responseWait := operationInfo{responseDiskExtend.ID, "DONE"}
	wait := mockClient.EXPECT().Send("operation.info",
		paramsWait, gomock.Any()).SetArg(2, responseWait).Return(nil).After(delete)

	paramsDiskInfo := []interface{}{responseDiskExtend.DiskID}
	responseDiskInfo := disks[0]
	responseDiskInfo.Name = newName
	mockClient.EXPECT().Send("hosting.disk.info",
		paramsDiskInfo, gomock.Any()).SetArg(2, responseDiskInfo).Return(nil).After(wait)

	diskparam := hosting.Disk{ID: strconv.Itoa(disks[0].ID)}
	disk, _ := testHosting.RenameDisk(diskparam, newName)
	if disk.Name != newName {
		t.Errorf("Error, expected disk name to be %s, got '%s' instead",
			newName, disk.Name)
	}
}

func TestDeleteDiskBadID(t *testing.T) {
	cl, _ := client.NewClientv4("", "1234")
	testHosting := Newv4Hosting(cl)

	disk := hosting.Disk{
		ID: "ThisisnotAnID",
	}
	err := testHosting.DeleteDisk(disk)
	if err == nil {
		t.Errorf("Error, expected error when parsing ID")
	}
}

func TestCreateDiskBadRegionID(t *testing.T) {
	cl, _ := client.NewClientv4("", "1234")
	testHosting := Newv4Hosting(cl)

	diskspec := hosting.DiskSpec{
		RegionID: "ThisisnotAnID",
	}
	_, err := testHosting.CreateDisk(diskspec)
	if err == nil {
		t.Errorf("Error, expected error when parsing ID")
	}
}

func TestFilterDisksBadID(t *testing.T) {
	cl, _ := client.NewClientv4("", "1234")
	testHosting := Newv4Hosting(cl)

	filter := hosting.DiskFilter{
		ID: "ThisisnotAnID",
	}
	_, err := testHosting.ListDisks(filter)
	if err == nil {
		t.Errorf("Error, expected error when parsing ID")
	}
}

func TestFilterDisksBadRegionID(t *testing.T) {
	cl, _ := client.NewClientv4("", "1234")
	testHosting := Newv4Hosting(cl)

	filter := hosting.DiskFilter{
		RegionID: "ThisisnotAnID",
	}
	_, err := testHosting.ListDisks(filter)
	if err == nil {
		t.Errorf("Error, expected error when parsing ID")
	}
}

func TestFilterDisksBadVMID(t *testing.T) {
	cl, _ := client.NewClientv4("", "1234")
	testHosting := Newv4Hosting(cl)

	filter := hosting.DiskFilter{
		VMID: "ThisisnotAnID",
	}
	_, err := testHosting.ListDisks(filter)
	if err == nil {
		t.Errorf("Error, expected error when parsing ID")
	}
}

func TestDiskFunctionsBadID(t *testing.T) {
	cl, _ := client.NewClientv4("", "1234")
	testHosting := Newv4Hosting(cl)
	
	disk := hosting.Disk{}

	_, err := testHosting.RenameDisk(disk, "Nom")
	
	if (err.(*HostingError).Err != ErrNotProvided) {
		t.Errorf("Error, Null ID expected in RenameDisk")
	}
	
	_, err = testHosting.ExtendDisk(disk, 1)
	if (err.(*HostingError).Err != ErrNotProvided) {
		t.Errorf("Error, Null ID expected in ExtendDisk")
	}
	
	err = testHosting.DeleteDisk(disk)
	if (err.(*HostingError).Err != ErrNotProvided) {
		t.Errorf("Error, Null ID expected in DeleteDisk")
	}

	err = testHosting.DeleteDisk(disk)
	if (err.(*HostingError).Err != ErrNotProvided) {
		t.Errorf("Error, Null ID expected in DeleteDisk")
	}

	disk.ID = "badid"
	_, err = testHosting.RenameDisk(disk, "nom")

	if (err.(*HostingError).Err != ErrParse) {
		t.Errorf("Error, Disk.ID must be a string of an integer in RenameDisk")
	}

	_, err = testHosting.ExtendDisk(disk, 1)
	if (err.(*HostingError).Err != ErrParse) {
		t.Errorf("Error, Disk.ID must be a string of an integer in ExtendDisk")
	}
	
	diskimg := hosting.DiskImage{RegionID:"badid"}
	_, err = testHosting.CreateDiskFromImage(hosting.DiskSpec{}, diskimg)
	if (err.(*HostingError).Err != ErrNotProvided) {
		t.Errorf("Error, DiskImage.DiskID must be a string of an integer in CreateDiskFromImage")
	}
	
	diskimg.DiskID = "666"
	
	_, err = testHosting.CreateDiskFromImage(hosting.DiskSpec{RegionID:"1"}, diskimg)
	if (err.(*HostingError).Err != ErrMismatch) {
		t.Errorf("Error, DiskImage.RegionID and DiskSpec.RegionID must be equal in CreateDiskFromImage")
	}
	
	_, err = testHosting.CreateDiskFromImage(hosting.DiskSpec{RegionID:"badid"}, diskimg)
	if (err.(*HostingError).Err != ErrParse) {
		t.Errorf("Error, DiskSpec.RegionID must be a string of an integer in CreateDiskFromImage")
	}
	
	diskimg.RegionID = "1"
	diskimg.DiskID = "badid"
	
	_, err = testHosting.CreateDiskFromImage(hosting.DiskSpec{RegionID:"1"}, diskimg)
	if (err.(*HostingError).Err != ErrParse) {
		t.Errorf("Error, DiskImage.DiskID must be a string of an integer in CreateDiskFromImage")
	}

}
