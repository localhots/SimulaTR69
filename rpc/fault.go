//nolint:revive
package rpc

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=FaultCode -trimprefix=Fault -output=fault_string.go
type FaultCode int

const (
	FaultMethodNotSupported FaultCode = 9000
	FaultRequestDenied      FaultCode = 9001
	FaultInternalError      FaultCode = 9002
	FaultInvalidArguments   FaultCode = 9003
	// When used in association with SetParameterValues, this must not be used
	// to indicate parameters in error.
	FaultResourcesExceeded FaultCode = 9004
	// Associated with SetParameterValues, GetParameterValues,
	// GetParameterNames, SetParameterAttributes, GetParameterAttributes,
	// AddObject, and DeleteObject.
	FaultInvalidParameterName FaultCode = 9005
	// Associated with SetParameterValues.
	FaultInvalidParameterType FaultCode = 9006
	// Associated with SetParameterValues.
	FaultInvalidParameterValue FaultCode = 9007
	// Attempt to set a non-writable parameter. Associated with
	// SetParameterValues.
	FaultNonWritableParameter FaultCode = 9008
	// Associated with SetParameterAttributes method.
	FaultNotificationRequestRejected FaultCode = 9009
	// Associated with Download, TransferComplete, or AutonomousTransferComplete
	// methods.
	FaultDownloadFailure FaultCode = 9010
	// Associated with Upload, TransferComplete, or AutonomousTransferComplete
	// methods.
	FaultUploadFailure FaultCode = 9011
	// File transfer server authentication failure. Associated with Upload,
	// Download, TransferComplete, or AutonomousTransferComplete methods.
	FaultFileTransferAuthenticationFailure FaultCode = 9012
	// Unsupported protocol for file transfer. Associated with Upload and
	// Download methods.
	FaultFileTransferUnsupportedProtocol FaultCode = 9013
	// Unable to join multicast group. Associated with Download,
	// TransferComplete, or AutonomousTransferComplete methods.
	FaultDownloadFailureJoinMulticastGroup FaultCode = 9014
	// Unable to contact file server. Associated with Download,
	// TransferComplete, or AutonomousTransferComplete methods.
	FaultDownloadFailureContactFileServer FaultCode = 9015
	// Unable to access file. Associated with Download, TransferComplete, or
	// AutonomousTransferComplete methods.
	FaultDownloadFailureAccessFile FaultCode = 9016
	// Unable to complete download. Associated with Download, TransferComplete,
	// or AutonomousTransferComplete methods.
	FaultDownloadFailureCompleteDownload FaultCode = 9017
	// Associated with Download, TransferComplete, or
	// AutonomousTransferComplete methods.
	FaultDownloadFailureFileCorrupted FaultCode = 9018
	// file authentication failure. Associated with Download, TransferComplete,
	// or AutonomousTransferComplete methods.
	FaultDownloadFailureAuthenticationFailure FaultCode = 9019

	FaultACSMethodNotSupported FaultCode = 8000
	FaultACSRequestDenied      FaultCode = 8001
	FaultACSInternalError      FaultCode = 8002
	FaultACSInvalidArguments   FaultCode = 8003
	FaultACSResoucesExceeded   FaultCode = 8004
	FaultACSRetryRequest       FaultCode = 8005
)

func (f FaultCode) Ptr() *FaultCode {
	return &f
}
