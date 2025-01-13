// Package rpc implements the TR-069 RPC (Remote Procedure Call) protocol,
// which is used for communication between Customer Premises Equipment (CPE)
// and Auto Configuration Servers (ACS). This package provides encoding and
// decoding functionalities for various TR-069 RPC messages, including
// Inform, GetRPCMethodsResponse, SetParameterValuesResponse, and more.
// It ensures proper XML formatting and namespace handling as per the TR-069
// specifications. The package also includes support for generating fault
// responses and encoding messages with optional pretty-printing.
package rpc
