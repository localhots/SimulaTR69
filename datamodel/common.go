package datamodel

import (
	"strconv"
	"strings"
	"time"
)

// DeviceID contains basic CPE info.
type DeviceID struct {
	Manufacturer string
	OUI          string
	ProductClass string
	SerialNumber string
}

// DeviceID returns a DeviceID populated from the datamodel.
func (dm *DataModel) DeviceID() DeviceID {
	return DeviceID{
		Manufacturer: dm.firstValue(
			"DeviceID.Manufacturer",
			"Device.DeviceInfo.Manufacturer",
			"InternetGatewayDevice.DeviceInfo.Manufacturer",
		),
		OUI: dm.firstValue(
			"DeviceID.OUI",
			"Device.DeviceInfo.ManufacturerOUI",
			"InternetGatewayDevice.DeviceInfo.ManufacturerOUI",
		),
		ProductClass: dm.firstValue(
			"DeviceID.ProductClass",
			"Device.DeviceInfo.ProductClass",
			"InternetGatewayDevice.DeviceInfo.ProductClass",
		),
		SerialNumber: dm.firstValue(
			"DeviceID.SerialNumber",
			"Device.DeviceInfo.SerialNumber",
			"InternetGatewayDevice.DeviceInfo.SerialNumber",
		),
	}
}

const (
	pathSerialNumber                = "DeviceInfo.SerialNumber"
	pathSoftwareVersion             = "DeviceInfo.SoftwareVersion"
	pathUptime                      = "DeviceInfo.UpTime"
	pathConnectionRequestURL        = "ManagementServer.ConnectionRequestURL"
	pathUDPConnectionRequestAddress = "ManagementServer.UDPConnectionRequestAddress"
	pathParameterKey                = "ManagementServer.ParameterKey"
	pathPeriodicInformEnable        = "ManagementServer.PeriodicInformEnable"
	pathPeriodicInformTime          = "ManagementServer.PeriodicInformTime"
	pathPeriodicInformInterval      = "ManagementServer.PeriodicInformInterval"
)

// SetSerialNumber sets serial number to the given value.
func (dm *DataModel) SetSerialNumber(val string) {
	dm.SetValue(pathSerialNumber, val)
}

// ConnectionRequestURL returns the connection request URL.
func (dm *DataModel) ConnectionRequestURL() Parameter {
	p, _ := dm.GetValue(pathConnectionRequestURL)
	return p
}

// SetConnectionRequestURL sets connection request URL to the given value.
func (dm *DataModel) SetConnectionRequestURL(val string) {
	dm.SetValue(pathConnectionRequestURL, val)
}

// UDPConnectionRequestAddress returns the UDP connection request address.
func (dm *DataModel) UDPConnectionRequestAddress() Parameter {
	p, _ := dm.GetValue(pathUDPConnectionRequestAddress)
	return p
}

// SetConnectionRequestURL sets UDP connection request address to the given value.
func (dm *DataModel) SetUDPConnectionRequestAddress(val string) {
	dm.SetValue(pathUDPConnectionRequestAddress, val)
}

// SetParameterKey sets parameter key to the given value.
func (dm *DataModel) SetParameterKey(val string) {
	dm.SetValue(pathParameterKey, val)
}

// PeriodicInformEnabled returns true if periodic inform is enabled.
func (dm *DataModel) PeriodicInformEnabled() bool {
	p, ok := dm.GetValue(pathPeriodicInformEnable)
	if !ok {
		return false
	}
	b, _ := strconv.ParseBool(p.Value)
	return b
}

// PeriodicInformInterval returns the value of periodic inform interval.
func (dm *DataModel) PeriodicInformInterval() time.Duration {
	const defaultInterval = 5 * time.Minute
	const secondsInDay = int64(24 * time.Hour / time.Second)
	p, ok := dm.GetValue(pathPeriodicInformInterval)
	if !ok {
		return defaultInterval
	}
	i, err := strconv.ParseInt(p.Value, 10, 32)
	if err != nil || i == 0 || i > secondsInDay {
		return defaultInterval
	}
	return time.Duration(i) * time.Second
}

// SetPeriodicInformInterval sets periodic inform interval to the given value.
func (dm *DataModel) SetPeriodicInformInterval(sec int64) {
	dm.SetValue(pathPeriodicInformInterval, strconv.FormatInt(sec, 10))
}

// PeriodicInformTime returns the value of periodic inform time.
func (dm *DataModel) PeriodicInformTime() time.Time {
	p, ok := dm.GetValue(pathPeriodicInformTime)
	if !ok {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, p.Value)
	if err != nil {
		return time.Time{}
	}
	return t
}

// SetPeriodicInformTime sets periodic inform time to the given value.
func (dm *DataModel) SetPeriodicInformTime(ts time.Time) {
	dm.SetValue(pathPeriodicInformTime, ts.UTC().Format(time.RFC3339))
}

// IsPeriodicInformParameter returns true if periodic inform is configured.
func (dm *DataModel) IsPeriodicInformParameter(name string) bool {
	if strings.HasSuffix(name, pathPeriodicInformInterval) {
		return true
	}
	if strings.HasSuffix(name, pathPeriodicInformTime) {
		return true
	}
	if strings.HasSuffix(name, pathPeriodicInformEnable) {
		return true
	}
	return false
}

// SetFirmwareVersion sets the new firmware version value.
func (dm *DataModel) SetFirmwareVersion(ver string) {
	dm.SetValue(pathSoftwareVersion, ver)
}

func (dm *DataModel) SetUptime(dur time.Duration) {
	dm.SetValue(pathUptime, strconv.Itoa(int(dur/time.Second)))
}
