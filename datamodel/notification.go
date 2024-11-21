package datamodel

import (
	"github.com/localhots/SimulaTR69/rpc"
)

// NotifyParams returns a list of parameters that should be included in the next
// inform message. This will always include forced parameters.
func (dm *DataModel) NotifyParams() []string {
	params := dm.ForcedInformParameters()
	dm.values.forEach(func(p Parameter) (cont bool) {
		if p.Notification == rpc.AttributeNotificationPassive {
			params = append(params, p.Path)
		}
		return true
	})

	return params
}

// ForcedInformParameters values that must be on every inform, according to the
// datamodel specifications.
//
// TR-098: https://cwmp-data-models.broadband-forum.org/tr-098-1-2-0.html#forced-inform-parameters
// TR-181: https://cwmp-data-models.broadband-forum.org/tr-181-2-18-1-cwmp.html#forced-inform-parameters
func (dm *DataModel) ForcedInformParameters() []string {
	common := []string{
		"DeviceInfo.HardwareVersion",
		"DeviceInfo.SoftwareVersion",
		"DeviceInfo.ProvisioningCode",
		"ManagementServer.ParameterKey",
		"ManagementServer.ConnectionRequestURL",
	}
	switch dm.version {
	case tr098:
		return append(common,
			"DeviceSummary",
			"DeviceInfo.SpecVersion",
			"DeviceInfo.ProvisioningCode",
		)
	case tr181:
		return append(common,
			"RootDataModelVersion",
			"ManagementServer.AliasBasedAddressing",
		)
	default:
		return common
	}
}
