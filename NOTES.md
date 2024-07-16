## Events
```
0 BOOSTRAP
1 BOOT
2 PERIODIC
3 SCHEDULED
4 VALUE CHANGE
5 KICKED
6 CONNECTION REQUEST
7 TRANSFER COMPLETE
8 DIAGNOSTICS COMPLETE
9 REQUEST DOWNLOAD
10 AUTONOMOUS TRANSFER COMPLETE
M Reboot
M ScheduleInform
M Download
M Upload
```

## ACS methods
```
Inform
  Request:
    DeviceId DeviceIdStruct
    Event EventStruct[64]
    MaxEnvelopes unsignedInt must always be 1
    CurrentTime dateTime
    RetryCount unsignedInt Number of prior times an attempt was made to retry this session.
    ParameterList ParameterValueStruct[]
  Response:
    MaxEnvelopes unsignedInt must always be 1
GetRPCMethods
TransferComplete
  CommandKey string(32)
  FaultStruct FaultStruct
  StartTime dateTime
  CompleteTime dateTime
```
## CPE methods
```
GetRPCMethods
  Response:
    MethodList string(64)[]
SetParameterValues
  Request:
    ParameterList ParameterValueStruct[]
    ParameterKey string(32)
  Response:
    Status int[0:1]
      0 = All Parameter changes have been validated and applied.
      1 = All Parameter changes have been validated and committed, but some or
          all are not yet applied (for example, if a reboot is required before
          the new values are applied)
GetParameterValues
  Request:
    ParameterNames string(256)[]
  Response:
    ParameterList ParameterValueStruct[]
GetParameterNames
  Request:
    ParameterPath string(256)
    NextLevel boolean
  Response:
    ParameterList ParameterInfoStruct[]
SetParameterAttributes
  Request:
    ParameterList SetParameterAttributesStruct[]
  Response: no arguments
GetParameterAttributes
  Request:
    ParameterNames string(256)[]
  Response:
    ParameterList ParameterAttributeStruct[]
AddObject
  Request:
    ObjectName string(256) Must end with a dot: Top.Group.Object.
    ParameterKey string(32)
      The value of this argument is left to the
      discretion of the ACS, and MAY be left empty.
  Response:
    InstanceNumber UnsignedInt[1:]
    Status int[0:1]
      0 = The object has been created.
      1 = The object creation has been validated and committed, but not yet applied
      (for example, if a reboot is required before the new object can be applied).
DeleteObject
  Request:
    ObjectName string(256) Must end with a dot: Top.Group.Object.
    ParameterKey string(32)
      The value of this argument is left to the
      discretion of the ACS, and MAY be left empty.
  Response:
    Status int[0:1] A successful response to this method returns an integer enumeration defined as follows:
      0 = The object has been deleted.
      1 = The object deletion has been validated and committed, but not yet applied (for example, if a
      reboot is required before the object can be deleted).
Reboot
  Request:
    CommandKey string(32) up to ACS
Download
  Request:
    CommandKey string(32) Up to ACS
    FileType string(64)
      "1 Firmware Upgrade Image" (must be supported)
      "2 Web Content"
      “3 Vendor Configuration File”
      "X <OUI> <Vendor-specific identifier>"
    URL string(256)
    Username string(256) Basic auth??
    Password string(256)
    FileSize unsignedInt
    TargetFileName string(256)
    DelaySeconds unsignedInt
    SuccessURL string(256) the CPE SHOULD redirect the user’s browser to if the download completes successfully
    FailureURL string(256)
  Response:
    Status int[0:1]
      0 = Download has completed and been applied.
      1 = Download has not yet been completed and applied (for example, if the CPE needs
      to reboot itself before it can perform the file download, or if the CPE needs to reboot
      itself before it can apply the downloaded file).
    StartTime dateTime
    CompleteTime dateTime
FactoryReset
```
## Vendor methods
```
X_{OUI}_{MethodName}
```
## Structures
```
ParameterValueStruct
  Name string(256)
  Value anySimpleType

ParameterInfoStruct
  Name string(256)
  Writable boolean

SetParameterAttributesStruct
  Name string(256)
  NotificationChange boolean
  Notification int[0:2]
    0 = Notification off. The CPE need not inform the
    ACS of a change to the specified parameter(s).
    1 = Passive notification. Whenever the specified
    parameter value changes, the CPE MUST
    include the new value in the ParameterList in the
    Inform message that is sent the next time a
    session is established to the ACS.
    If the CPE has rebooted, or the URL of the ACS
    has changed since the last session, the CPE
    MAY choose not to include the list of changed
    parameters in the first session established with
    the new ACS.
    2 = Active notification. Whenever the specified
    parameter value changes, the CPE MUST initiate
    a session to the ACS, and include the new value
    in the ParameterList in the associated Inform
    message.
  AccessListChange boolean
  AccessList string(64)[]
    Only Subscriber value is supported
    “Subscriber” Indicates write access by an
    interface controlled on the
    subscriber LAN. Includes any
    and all such LAN-side
    mechanisms, which MAY include
    but are not limited to TR-064
    (LAN-side DSL CPE
    Configuration Protocol), UPnP,
    the device’s user interface, client-
    side telnet, and client-side
    SNMP.

ParameterAttributeStruct
  Name string(256)
  Notification int[0:2]
    0 = Notification off. The CPE need not inform the
    ACS of a change to the specified parameter(s).
    1 = Passive notification. Whenever the specified
    parameter value changes, the CPE MUST
    include the new value in the ParameterList in
    the Inform message that is sent the next time a
    session is established to the ACS.
    2 = Active notification. Whenever the specified
    parameter value changes, the CPE MUST
    initiate a session to the ACS, and include the
    new value in the ParameterList in the
    associated Inform message.
  AccessList string(64)[]
    Only Subscriber value is supported
    “Subscriber” Indicates write access by an
    interface controlled on the
    subscriber LAN. Includes any
    and all such LAN-side
    mechanisms, which MAY include
    but are not limited to TR-064
    (LAN-side DSL CPE
    Configuration Protocol), UPnP,
    the device’s user interface, client-
    side telnet, and client-side
    SNMP.

DeviceIdStruct
  Manufacturer string(64)
  OUI string(6) six hex digits uppercase
  ProductClass string(64)
  SerialNumber string(64)

EventStruct
  EventCode string(64) See Table 7 in section 3.7.1.5 for event codes
  CommandKey string(32)
    ScheduledInform method (EventCode = “M ScheduleInform”)
    Reboot method (EventCode = “M Reboot”)
    Download method (EventCode = “M Download”)
    Upload method (EventCode = “M Upload”)
    For all other EventCode values defined in this specification, the value of CommandKey MUST be an empty string.

FaultStruct
  FaultCode unsignedInt The numerical fault code as defined in section A.5.1.
    In the case of a fault, allowed values are:
    9001, 9002, 9010, 9011, 9012, 9014, 9015, 9016, 9017, 9018, 9019.
    Avalue of 0 (zero) indicates no fault.
  FaultString string(256) A human-readable text description of the fault
    This field SHOULD be empty if the FaultCode equals 0 (zero).
```
