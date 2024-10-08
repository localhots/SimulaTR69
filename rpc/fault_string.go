// Code generated by "stringer -type=FaultCode -trimprefix=Fault -output=fault_string.go"; DO NOT EDIT.

package rpc

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[FaultMethodNotSupported-9000]
	_ = x[FaultRequestDenied-9001]
	_ = x[FaultInternalError-9002]
	_ = x[FaultInvalidArguments-9003]
	_ = x[FaultResourcesExceeded-9004]
	_ = x[FaultInvalidParameterName-9005]
	_ = x[FaultInvalidParameterType-9006]
	_ = x[FaultInvalidParameterValue-9007]
	_ = x[FaultNonWritableParameter-9008]
	_ = x[FaultNotificationRequestRejected-9009]
	_ = x[FaultDownloadFailure-9010]
	_ = x[FaultUploadFailure-9011]
	_ = x[FaultFileTransferAuthenticationFailure-9012]
	_ = x[FaultFileTransferUnsupportedProtocol-9013]
	_ = x[FaultDownloadFailureJoinMulticastGroup-9014]
	_ = x[FaultDownloadFailureContactFileServer-9015]
	_ = x[FaultDownloadFailureAccessFile-9016]
	_ = x[FaultDownloadFailureCompleteDownload-9017]
	_ = x[FaultDownloadFailureFileCorrupted-9018]
	_ = x[FaultDownloadFailureAuthenticationFailure-9019]
	_ = x[FaultACSMethodNotSupported-8000]
	_ = x[FaultACSRequestDenied-8001]
	_ = x[FaultACSInternalError-8002]
	_ = x[FaultACSInvalidArguments-8003]
	_ = x[FaultACSResoucesExceeded-8004]
	_ = x[FaultACSRetryRequest-8005]
}

const (
	_FaultCode_name_0 = "ACSMethodNotSupportedACSRequestDeniedACSInternalErrorACSInvalidArgumentsACSResoucesExceededACSRetryRequest"
	_FaultCode_name_1 = "MethodNotSupportedRequestDeniedInternalErrorInvalidArgumentsResourcesExceededInvalidParameterNameInvalidParameterTypeInvalidParameterValueNonWritableParameterNotificationRequestRejectedDownloadFailureUploadFailureFileTransferAuthenticationFailureFileTransferUnsupportedProtocolDownloadFailureJoinMulticastGroupDownloadFailureContactFileServerDownloadFailureAccessFileDownloadFailureCompleteDownloadDownloadFailureFileCorruptedDownloadFailureAuthenticationFailure"
)

var (
	_FaultCode_index_0 = [...]uint8{0, 21, 37, 53, 72, 91, 106}
	_FaultCode_index_1 = [...]uint16{0, 18, 31, 44, 60, 77, 97, 117, 138, 158, 185, 200, 213, 246, 277, 310, 342, 367, 398, 426, 462}
)

func (i FaultCode) String() string {
	switch {
	case 8000 <= i && i <= 8005:
		i -= 8000
		return _FaultCode_name_0[_FaultCode_index_0[i]:_FaultCode_index_0[i+1]]
	case 9000 <= i && i <= 9019:
		i -= 9000
		return _FaultCode_name_1[_FaultCode_index_1[i]:_FaultCode_index_1[i+1]]
	default:
		return "FaultCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
