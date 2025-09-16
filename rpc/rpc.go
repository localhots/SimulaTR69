//nolint:revive
package rpc

import (
	"fmt"
)

type (
	AttributeNotification int
)

const (
	NSEnc  = "http://schemas.xmlsoap.org/soap/encoding/"
	NSEnv  = "http://schemas.xmlsoap.org/soap/envelope/"
	NSXSD  = "http://www.w3.org/2001/XMLSchema"
	NSXSI  = "http://www.w3.org/2001/XMLSchema-instance"
	NSCWMP = "urn:dslforum-org:cwmp-1-0"

	EventBootstrap                  = "0 BOOTSTRAP"
	EventBoot                       = "1 BOOT"
	EventPeriodic                   = "2 PERIODIC"
	EventScheduled                  = "3 SCHEDULED"
	EventValueChange                = "4 VALUE CHANGE"
	EventKicked                     = "5 KICKED"
	EventConnectionRequest          = "6 CONNECTION REQUEST"
	EventTransferComplete           = "7 TRANSFER COMPLETE"
	EventDiagnosticsComplete        = "8 DIAGNOSTICS COMPLETE"
	EventRequestDownload            = "9 REQUEST DOWNLOAD"
	EventAutonomousTransferComplete = "10 AUTONOMOUS TRANSFER COMPLETE"
	EventReboot                     = "M Reboot"
	EventScheduleInform             = "M ScheduleInform"
	EventDownload                   = "M Download"
	EventUpload                     = "M Upload"

	FileTypeFirmwareUpgradeImage    = "1 Firmware Upgrade Image"
	FileTypeWebContent              = "2 Web Content"
	FileTypeVendorConfigurationFile = "3 Vendor Configuration File"

	// AttributeNotificationOff indicates that the CPE need not inform the ACS
	// of a change to the specified parameter(s).
	AttributeNotificationOff AttributeNotification = 0
	// AttributeNotificationPassive indicates that whenever the specified
	// parameter value changes, the CPE MUST include the new value in the
	// ParameterList in the Inform message that is sent the next time a session
	// is established to the ACS. If the CPE has rebooted, or the URL of the ACS
	// has changed since the last session, the CPE MAY choose not to include the
	// list of changed parameters in the first session established with the new
	// ACS.
	AttributeNotificationPassive AttributeNotification = 1
	// AttributeNotificationActive indicates that whenever the specified
	// parameter value changes, the CPE MUST initiate a session to the ACS, and
	// include the new value in the ParameterList in the associated Inform
	// message.
	AttributeNotificationActive AttributeNotification = 2

	// MaxEnvelopes MUST be set to a value of 1 because this version of the
	// protocol supports only a single envelope per message, and on reception
	// its value MUST be ignored.
	MaxEnvelopes = 1

	// Download has completed and been applied.
	DownloadCompleted = 0
	// Download has not yet been completed and applied (for example, if the CPE
	// needs to reboot itself before it can perform the file download, or if the
	// CPE needs to reboot itself before it can apply the downloaded file).
	DownloadNotCompleted = 1
)

type EventStruct struct {
	EventCode  string
	CommandKey string
}

type DeviceID struct {
	Manufacturer string
	OUI          string
	ProductClass string
	SerialNumber string
}

type ParameterAttribute struct {
	Name               string
	NotificationChange bool
	Notification       int
	AccessListChange   bool
	AccessList         []string
}

func SupportedMethods() []string {
	return []string{
		"GetRPCMethods",
		"SetParameterValues",
		"GetParameterValues",
		"GetParameterNames",
		"SetParameterAttributes",
		"GetParameterAttributes",
		"AddObject",
		"DeleteObject",
		"Reboot",
		"Download",
		"FactoryReset",
	}
}

func ArrayType(typ string, size int) string {
	return fmt.Sprintf("%s[%d]", typ, size)
}
