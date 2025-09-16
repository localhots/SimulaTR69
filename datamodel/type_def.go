package datamodel

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/localhots/SimulaTR69/rpc"
)

type typeDef struct {
	name     string
	min, max *int
}

// Not covered:
// * steps, e.g. int(0:100 step 5)
// * only mins? e.g. int(5:)
// * floats?
// * enums, e.g. int(1,2,3)
// * string enums with multiple-word values, e.g. string(foo,"two words",bar)
// * something else?
var typeDefRegex = regexp.MustCompile(`^(?:xsd:)?(?P<name>\w+)(\(((?P<min>\d+):)?(?P<max>\d+)\))?$`)

func parseTypeDef(str string) (*typeDef, error) {
	m := typeDefRegex.FindStringSubmatch(strings.TrimSpace(str))
	if len(m) == 0 {
		return nil, errors.New("invalid type definition")
	}

	var td typeDef
	for i, name := range typeDefRegex.SubexpNames() {
		if name == "" || len(m) <= i {
			continue
		}

		val := m[i]
		switch {
		case name == "name":
			td.name = val
		case name == "min" && val != "":
			minVal, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("parse type min: %w", err)
			}
			td.min = &minVal
		case name == "max" && val != "":
			maxVal, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("parse type max: %w", err)
			}
			td.max = &maxVal
		}
	}
	return td.normalize(), nil
}

func (td *typeDef) String() string {
	if td.min != nil {
		return fmt.Sprintf("%s(%d:%d)", rpc.XSD(td.name), *td.min, *td.max)
	}
	if td.max != nil {
		return fmt.Sprintf("%s(%d)", rpc.XSD(td.name), *td.max)
	}
	return rpc.XSD(td.name)
}

func (td *typeDef) normalize() *typeDef {
	switch td.name {
	case rpc.TypeBase64,
		rpc.TypeBoolean,
		rpc.TypeDateTime,
		rpc.TypeHEXBinary,
		rpc.TypeInt,
		rpc.TypeString,
		rpc.TypeUnsignedInt,
		rpc.TypeUnsignedLong,
		rpc.TypeIPAddress,
		rpc.TypeIPPrefix,
		rpc.TypeIPv4Address,
		rpc.TypeIPv6Address,
		rpc.TypeIPv6Prefix,
		rpc.TypeMACAddress,
		rpc.TypeGenerator:
		// Supported type, nothing to normalize
	case "long":
		td.name = rpc.TypeUnsignedLong
	case rpc.TypeBase64Binary:
		// Not part of TR-181 spec but is widely used
		td.name = rpc.TypeBase64Binary
	default:
		// Fallback to string
		td.name = rpc.TypeString
	}
	return td
}
