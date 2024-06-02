package diskie

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

type Conn struct {
	conn *dbus.Conn
}

type BlockDevice struct {
	ObjectPath            string
	Device                *string
	PreferredDevice       *string
	Symlinks              *[]string
	DeviceNumber          *uint64
	Id                    *string
	Size                  *uint64
	ReadOnly              *bool
	Drive                 *Drive
	IdUsage               *string
	IdType                *string
	IdVersion             *string
	IdLabel               *string
	IdUUID                *string
	CryptoBackingDevice   *string
	HintPartitionable     *bool
	HintSystem            *bool
	HintIgnore            *bool
	HintAuto              *bool
	HintName              *string
	HintIconName          *string
	HintSymbolicIconName  *string
	UserspaceMountOptions *[]string
	Partition             *Partition
	Filesystem            *Filesystem
	Encrypted             *Encrypted

	// custom attributes
	RootDrive     *Drive
	RootDevice    string
	PreferredSize *uint64
}

type Drive struct {
	Vendor                *string
	Model                 *string
	Revision              *string
	Serial                *string
	WWN                   *string
	Id                    *string
	Media                 *string
	MediaCompatibility    *[]string
	MediaRemovable        *bool
	MediaAvailable        *bool
	MediaChangeDetected   *bool
	Size                  *uint64
	TimeDetected          *uint64
	TimeMediaDetected     *uint64
	Optical               *bool
	OpticalBlank          *bool
	OpticalNumTracks      *uint32
	OpticalNumAudioTracks *uint32
	OpticalNumDataTracks  *uint32
	OpticalNumSessions    *uint32
	RotationRate          *int32
	ConnectionBus         *string
	Seat                  *string
	Removable             *bool
	Ejectable             *bool
	SortKey               *string
	CanPowerOff           *bool
	SiblingId             *string
}

type Partition struct {
	Number      *uint32
	Type        *string
	Flags       *uint64
	Offset      *uint64
	Size        *uint64
	Name        *string
	UUID        *string
	IsContainer *bool
	IsContained *bool
}

type Filesystem struct {
	MountPoints *[]string
	Size        *uint64
}

type Encrypted struct {
	HintEncryptionType *string
	MetadataSize       *uint64
	CleartextDevice    *string
}

func Connect() (*Conn, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("could not connect to system bus: %w", err)
	}
	return &Conn{
		conn: conn,
	}, nil
}

func (c *Conn) BlockDevices() (*BlockMap, error) {
	obj := c.conn.Object(
		"org.freedesktop.UDisks2",
		"/org/freedesktop/UDisks2/Manager",
	)

	method := "org.freedesktop.UDisks2.Manager.GetBlockDevices"

	var paths []dbus.ObjectPath

	err := obj.Call(method, 0, map[string]interface{}{}).Store(&paths)
	if err != nil {
		return nil, fmt.Errorf("method %s failed: %w", method, err)
	}

	blockmap := make(map[string]*BlockDevice, len(paths))

	for _, path := range paths {

		obj := c.conn.Object("org.freedesktop.UDisks2", path)
		property := "org.freedesktop.UDisks2.Block"

		var store map[string]dbus.Variant

		err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, property).Store(&store)
		if err != nil {
			return nil, fmt.Errorf("could not get property %s: %w", property, err)
		}

		block := BlockDevice{
			ObjectPath: string(path),
		}

		for k, v := range store {
			switch k {
			case "Device":
				val := toString(v.Value().([]byte))
				block.Device = &val
			case "PreferredDevice":
				val := toString(v.Value().([]byte))
				block.PreferredDevice = &val
			case "Symlinks":
				val := []string{}
				for _, s := range v.Value().([][]byte) {
					val = append(val, toString(s))
				}
				block.Symlinks = &val
			case "DeviceNumber":
				val := v.Value().(uint64)
				block.DeviceNumber = &val
			case "Id":
				val := v.Value().(string)
				block.Id = &val
			case "Size":
				val := v.Value().(uint64)
				block.Size = &val
			case "ReadOnly":
				val := v.Value().(bool)
				block.ReadOnly = &val
			case "IdUsage":
				val := v.Value().(string)
				block.IdUsage = &val
			case "IdType":
				val := v.Value().(string)
				block.IdType = &val
			case "IdVersion":
				val := v.Value().(string)
				block.IdVersion = &val
			case "IdLabel":
				val := v.Value().(string)
				block.IdLabel = &val
			case "IdUUID":
				val := v.Value().(string)
				block.IdUUID = &val
			case "CryptoBackingDevice":
				val := string(v.Value().(dbus.ObjectPath))
				block.CryptoBackingDevice = &val
			case "HintPartitionable":
				val := v.Value().(bool)
				block.HintPartitionable = &val
			case "HintSystem":
				val := v.Value().(bool)
				block.HintSystem = &val
			case "HintIgnore":
				val := v.Value().(bool)
				block.HintIgnore = &val
			case "HintAuto":
				val := v.Value().(bool)
				block.HintAuto = &val
			case "HintName":
				val := v.Value().(string)
				block.HintName = &val
			case "HintIconName":
				val := v.Value().(string)
				block.HintIconName = &val
			case "HintSymbolicIconName":
				val := v.Value().(string)
				block.HintSymbolicIconName = &val
			case "UserspaceMountOptions":
				val := v.Value().([]string)
				block.UserspaceMountOptions = &val
			}
		}

		// BlockDevice.Drive
		var drivePath dbus.ObjectPath
		property = "org.freedesktop.UDisks2.Block.Drive"
		err = obj.StoreProperty(property, &drivePath)
		if err != nil {
			return nil, fmt.Errorf("could not get property %s: %w", property, err)
		}
		drive, err := c.getDrive(drivePath)
		if err != nil {
			return nil, fmt.Errorf("could not get BlockDevice.Drive: %w", err)
		}
		block.Drive = drive

		// BlockDevice.Partition
		partition, err := getPartition(obj)
		if err != nil {
			return nil, fmt.Errorf("could not get BlockDevice.Partition: %w", err)
		}
		block.Partition = partition

		// BlockDevice.Filesystem
		fs, err := getFilesystem(obj)
		if err != nil {
			return nil, fmt.Errorf("could not get BlockDevice.Filesystem: %w", err)
		}
		block.Filesystem = fs

		// BlockDevice.PreferredSize
		if fs != nil && fs.Size != nil && *fs.Size != 0 {
			block.PreferredSize = fs.Size
		} else if partition != nil && partition.Size != nil && *partition.Size != 0 {
			block.PreferredSize = partition.Size
		} else {
			block.PreferredSize = block.Size
		}

		// BlockDevice.Encrypted
		enc, err := getEncrypted(obj)
		if err != nil {
			return nil, fmt.Errorf("could not get BlockDevic.Encrypted: %w", err)
		}
		block.Encrypted = enc

		blockmap[block.ObjectPath] = &block
	}

	// BlockDevice.RootDrive
	var getRootDrive func(*BlockDevice) *Drive
	getRootDrive = func(b *BlockDevice) *Drive {
		if b.Drive != nil {
			return b.Drive
		}
		c := b.CryptoBackingDevice
		if c == nil || *c == "/" {
			return nil
		}
		b, has := blockmap[*c]
		if has {
			return getRootDrive(b)
		}
		panic(fmt.Errorf("CryptoBackingDevice not found in the list of devices: %s", *c))
	}
	for _, b := range blockmap {
		b.RootDrive = getRootDrive(b)
	}

	// BlockDevice.RootBackingDevice
	var getRootDevice func(*BlockDevice) string
	getRootDevice = func(b *BlockDevice) string {
		c := b.CryptoBackingDevice
		if c == nil || *c == "/" {
			return b.ObjectPath
		}
		b, has := blockmap[*c]
		if has {
			return getRootDevice(b)
		}
		panic(fmt.Errorf("CryptoBackingDevice not found in the list of devices: %s", *c))
	}
	for _, b := range blockmap {
		b.RootDevice = getRootDevice(b)
	}

	return &BlockMap{
		BlockMap: blockmap,
	}, nil
}

func (c *Conn) getDrive(path dbus.ObjectPath) (*Drive, error) {
	if path == "/" {
		return nil, nil
	}

	obj := c.conn.Object("org.freedesktop.UDisks2", path)
	property := "org.freedesktop.UDisks2.Drive"

	var store map[string]dbus.Variant

	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, property).Store(&store)
	if err != nil {
		return nil, fmt.Errorf("could not get property %s: %w", property, err)
	}

	var drive Drive

	for k, v := range store {
		switch k {
		case "Vendor":
			val := v.Value().(string)
			drive.Vendor = &val
		case "Model":
			val := v.Value().(string)
			drive.Model = &val
		case "Revision":
			val := v.Value().(string)
			drive.Revision = &val
		case "Serial":
			val := v.Value().(string)
			drive.Serial = &val
		case "WWN":
			val := v.Value().(string)
			drive.WWN = &val
		case "Id":
			val := v.Value().(string)
			drive.Id = &val
		case "Media":
			val := v.Value().(string)
			drive.Media = &val
		case "MediaCompatibility":
			val := v.Value().([]string)
			drive.MediaCompatibility = &val
		case "MediaRemovable":
			val := v.Value().(bool)
			drive.MediaRemovable = &val
		case "MediaAvailable":
			val := v.Value().(bool)
			drive.MediaAvailable = &val
		case "MediaChangeDetected":
			val := v.Value().(bool)
			drive.MediaChangeDetected = &val
		case "Size":
			val := v.Value().(uint64)
			drive.Size = &val
		case "TimeDetected":
			val := v.Value().(uint64)
			drive.TimeDetected = &val
		case "TimeMediaDetected":
			val := v.Value().(uint64)
			drive.TimeMediaDetected = &val
		case "Optical":
			val := v.Value().(bool)
			drive.Optical = &val
		case "OpticalBlank":
			val := v.Value().(bool)
			drive.OpticalBlank = &val
		case "OpticalNumTracks":
			val := v.Value().(uint32)
			drive.OpticalNumTracks = &val
		case "OpticalNumAudioTracks":
			val := v.Value().(uint32)
			drive.OpticalNumAudioTracks = &val
		case "OpticalNumDataTracks":
			val := v.Value().(uint32)
			drive.OpticalNumDataTracks = &val
		case "OpticalNumSessions":
			val := v.Value().(uint32)
			drive.OpticalNumSessions = &val
		case "RotationRate":
			val := v.Value().(int32)
			drive.RotationRate = &val
		case "ConnectionBus":
			val := v.Value().(string)
			drive.ConnectionBus = &val
		case "Seat":
			val := v.Value().(string)
			drive.Seat = &val
		case "Removable":
			val := v.Value().(bool)
			drive.Removable = &val
		case "Ejectable":
			val := v.Value().(bool)
			drive.Ejectable = &val
		case "SortKey":
			val := v.Value().(string)
			drive.SortKey = &val
		case "CanPowerOff":
			val := v.Value().(bool)
			drive.CanPowerOff = &val
		case "SiblingId":
			val := v.Value().(string)
			drive.SiblingId = &val
		}
	}

	return &drive, nil
}

func getEncrypted(obj dbus.BusObject) (*Encrypted, error) {
	property := "org.freedesktop.UDisks2.Encrypted"

	var store map[string]dbus.Variant

	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, property).Store(&store)

	if err != nil && strings.Contains(err.Error(), "No such interface") {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get property %s: %w", property, err)
	}

	var enc Encrypted

	for k, v := range store {
		switch k {
		case "HintEncryptionType":
			val := v.Value().(string)
			enc.HintEncryptionType = &val
		case "MetadataSize":
			val := v.Value().(uint64)
			enc.MetadataSize = &val
		case "CleartextDevice":
			val := string(v.Value().(dbus.ObjectPath))
			enc.CleartextDevice = &val
		}
	}

	return &enc, nil
}

func getFilesystem(obj dbus.BusObject) (*Filesystem, error) {
	property := "org.freedesktop.UDisks2.Filesystem"

	var store map[string]dbus.Variant

	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, property).Store(&store)

	if err != nil && strings.Contains(err.Error(), "No such interface") {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get property %s: %w", property, err)
	}

	var fs Filesystem

	for k, v := range store {
		switch k {
		case "MountPoints":
			val := []string{}
			for _, s := range v.Value().([][]byte) {
				val = append(val, toString(s))
			}
			fs.MountPoints = &val
		case "Size":
			val := v.Value().(uint64)
			fs.Size = &val
		}
	}

	return &fs, nil
}

func getPartition(obj dbus.BusObject) (*Partition, error) {
	property := "org.freedesktop.UDisks2.Partition"

	var store map[string]dbus.Variant

	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, property).Store(&store)

	if err != nil && strings.Contains(err.Error(), "No such interface") {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get property %s: %w", property, err)
	}

	var partition Partition

	for k, v := range store {
		switch k {
		case "Number":
			val := v.Value().(uint32)
			partition.Number = &val
		case "Type":
			val := v.Value().(string)
			partition.Type = &val
		case "Flags":
			val := v.Value().(uint64)
			partition.Flags = &val
		case "Offset":
			val := v.Value().(uint64)
			partition.Offset = &val
		case "Size":
			val := v.Value().(uint64)
			partition.Size = &val
		case "Name":
			val := v.Value().(string)
			partition.Name = &val
		case "UUID":
			val := v.Value().(string)
			partition.UUID = &val
		case "IsContainer":
			val := v.Value().(bool)
			partition.IsContainer = &val
		case "IsContained":
			val := v.Value().(bool)
			partition.IsContained = &val
		}
	}

	return &partition, nil
}

func toString(b []byte) string {
	return string(bytes.TrimRight(b, "\u0000"))
}
