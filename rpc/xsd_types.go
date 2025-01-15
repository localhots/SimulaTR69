package rpc

// List of supported types.
const (
	TypeObject       = "object"
	TypeBase64       = "base64"
	TypeBase64Binary = "base64Binary"
	TypeBoolean      = "boolean"
	TypeDateTime     = "dateTime"
	TypeHEXBinary    = "hexBinary"
	TypeInt          = "int"
	TypeLong         = "long"
	TypeString       = "string"
	TypeUnsignedInt  = "unsignedInt"
	TypeUnsignedLong = "unsignedLong"
	TypeIPAddress    = "IPAddress"
	TypeIPPrefix     = "IPPrefix"
	TypeIPv4Address  = "IPv4Address"
	TypeIPv6Address  = "IPv6Address"
	TypeIPv6Prefix   = "IPv6Prefix"
	TypeMACAddress   = "MACAddress"

	// TypeGenerator is a special type used to define a generator function.
	TypeGenerator = "sim:generator"
)

// XSD returns the XML Schema Definition (XSD) type for the given type.
func XSD(typ string) string {
	return "xsd:" + typ
}
