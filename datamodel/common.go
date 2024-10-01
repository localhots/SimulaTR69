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
	pathSerialNumber           = "DeviceInfo.SerialNumber"
	pathSoftwareVersion        = "DeviceInfo.SoftwareVersion"
	pathUptime                 = "DeviceInfo.UpTime"
	pathConnectionRequestURL   = "ManagementServer.ConnectionRequestURL"
	pathParameterKey           = "ManagementServer.ParameterKey"
	pathPeriodicInformEnable   = "ManagementServer.PeriodicInformEnable"
	pathPeriodicInformTime     = "ManagementServer.PeriodicInformTime"
	pathPeriodicInformInterval = "ManagementServer.PeriodicInformInterval"
)

// SetSerialNumber sets serial number to the given value.
func (dm *DataModel) SetSerialNumber(val string) {
	dm.SetValue(pathSerialNumber, val)
}

// ConnectionRequestURL returns the connection request URL.
func (dm *DataModel) ConnectionRequestURL() Parameter {
	return dm.GetValue(pathConnectionRequestURL)
}

// SetConnectionRequestURL sets connection request URL to the given value.
func (dm *DataModel) SetConnectionRequestURL(val string) {
	dm.SetValue(pathConnectionRequestURL, val)
}

// SetParameterKey sets parameter key to the given value.
func (dm *DataModel) SetParameterKey(val string) {
	dm.SetValue(pathParameterKey, val)
}

// PeriodicInformEnabled returns true if periodic inform is enabled.
func (dm *DataModel) PeriodicInformEnabled() bool {
	val := dm.GetValue(pathPeriodicInformEnable)
	b, _ := strconv.ParseBool(val.Value)
	return b
}

// PeriodicInformInterval returns the value of periodic inform interval.
func (dm *DataModel) PeriodicInformInterval() time.Duration {
	const defaultInterval = 5 * time.Minute
	const secondsInDay = int64(24 * time.Hour / time.Second)
	val := dm.GetValue(pathPeriodicInformInterval)
	i, _ := strconv.ParseInt(val.Value, 10, 32)
	if i == 0 || i > secondsInDay {
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
	val := dm.GetValue(pathPeriodicInformTime)
	i, _ := strconv.ParseInt(val.Value, 10, 32)
	return time.Unix(i, 0)
}

// SetPeriodicInformTime sets periodic inform time to the given value.
func (dm *DataModel) SetPeriodicInformTime(ts time.Time) {
	dm.SetValue(pathPeriodicInformTime, strconv.FormatInt(ts.Unix(), 10))
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
