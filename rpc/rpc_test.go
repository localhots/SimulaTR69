package rpc

import _ "embed"

var (
	//
	// Requests.
	//

	//go:embed test_data/get_rpc_methods_request.xml
	getRPCMethodsRequestTestData []byte

	//go:embed test_data/set_parameter_values_request.xml
	setParameterValuesRequestTestData []byte

	//go:embed test_data/get_parameter_values_request.xml
	getParameterValuesRequestTestData []byte

	//go:embed test_data/get_invalid_parameter_values_request.xml
	getInvalidParameterValuesRequestTestData []byte

	//go:embed test_data/get_parameter_names_request.xml
	getParameterNamesRequestTestData []byte

	//go:embed test_data/add_object_request.xml
	addObjectTestData []byte

	//go:embed test_data/delete_object_request.xml
	deleteObjectRequestTestData []byte

	//go:embed test_data/reboot_request.xml
	rebootRequestTestData []byte

	//go:embed test_data/download_request.xml
	downloadRequestTestData []byte

	//go:embed test_data/factory_reset_request.xml
	factoryResetRequestTestData []byte

	//go:embed test_data/inform_request.xml
	informRequestTestData []byte

	//go:embed test_data/set_parameter_attributes_request.xml
	setParameterAttributesRequestTestData []byte

	//go:embed test_data/get_parameter_attributes_request.xml
	getParameterAttributesRequestTestData []byte

	//go:embed test_data/transfer_complete_success_request.xml
	transferCompleteSuccessRequestTestData []byte

	//go:embed test_data/transfer_complete_fault_request.xml
	transferCompleteFaultRequestTestData []byte

	//go:embed test_data/autonomous_transfer_complete_request.xml
	autonomousTransferCompleteRequestTestData []byte

	//
	// Responses.
	//

	//go:embed test_data/inform_response.xml
	informResponseTestData []byte

	//go:embed test_data/get_rpc_methods_response.xml
	getRPCMethodsResponseTestData []byte

	//go:embed test_data/set_parameter_values_response.xml
	setParameterValuesResponseTestData []byte

	//go:embed test_data/get_parameter_values_response.xml
	getParameterValuesResponseTestData []byte

	//go:embed test_data/get_parameter_values_fault_response.xml
	getParameterValuesFaultResponseTestData []byte

	//go:embed test_data/get_parameter_names_response.xml
	getParameterNamesResponseTestData []byte

	//go:embed test_data/set_parameter_attributes_response.xml
	setParameterAttributesResponseTestData []byte

	//go:embed test_data/get_parameter_attributes_response.xml
	getParameterAttributesResponseTestData []byte

	//go:embed test_data/add_object_response.xml
	addObjectResponseTestData []byte

	//go:embed test_data/delete_object_response.xml
	deleteObjectResponseTestData []byte

	//go:embed test_data/reboot_response.xml
	rebootResponseTestData []byte

	//go:embed test_data/download_response.xml
	downloadResponseTestData []byte

	//go:embed test_data/factory_reset_response.xml
	factoryResetResponseTestData []byte

	//go:embed test_data/transfer_complete_response.xml
	transferCompleteResponseTestData []byte

	//go:embed test_data/autonomous_transfer_complete_response.xml
	autonomousTransferCompleteResponseTestData []byte

	//go:embed test_data/fault_response.xml
	faultResponseTestData []byte

	//go:embed test_data/fault_set_parameter_values_response.xml
	faultSetParameterValuesResponseTestData []byte
)
