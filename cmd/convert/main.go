// This package allows to convert datamodels into a CSV format that is supported
// by this simulator.
package main

import (
	"cmp"
	"encoding/csv"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/rs/zerolog/log"
)

var (
	source = flag.String("src", "", "Source file path")
	typ    = flag.String("type", "GetParameterValuesResponse", "Source file type. Supported types: GetParameterValuesResponse")
	dest   = flag.String("dest", "datamodel.csv", "Destination file path")
)

func main() {
	flag.Parse()
	if *source == "" {
		log.Fatal().Msg("Source file path is empty")
	}
	if *typ == "" {
		log.Fatal().Msg("Source file type is empty")
	}
	if *dest == "" {
		log.Fatal().Msg("Destination file path is empty")
	}

	switch *typ {
	case "GetParameterValuesResponse":
		pvr := read(*source)
		dm := convertGetParameterValuesResponse(pvr)
		if err := save(*dest, dm); err != nil {
			log.Fatal().Err(err).Msg("Failed to save datamodel file")
		}
	default:
		log.Fatal().Str("type", *typ).Msg("Unknown format")
	}
}

func read(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("Failed to read source file")
	}
	return b
}

func save(path string, params []datamodel.Parameter) error {
	fd, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create destination file: %s", path)
	}
	defer fd.Close()

	slices.SortFunc(params, func(a, b datamodel.Parameter) int {
		return cmp.Compare(strings.ToLower(a.Path), strings.ToLower(b.Path))
	})

	w := csv.NewWriter(fd)
	err = w.Write([]string{"Parameter", "Object", "Writable", "Value", "Type"})
	if err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, p := range params {
		err := w.Write([]string{p.Path, strconv.FormatBool(p.Object), strconv.FormatBool(p.Writable), p.Value, p.Type})
		if err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("csv write error: %w", err)
	}

	return nil
}

// FIXME: cyclo complexity is too high, rewrite it
// nolint:gocyclo,musttag
func convertGetParameterValuesResponse(b []byte) []datamodel.Parameter {
	var gpv struct {
		XMLName       xml.Name `xml:"GetParameterValuesResponse"`
		ParameterList struct {
			ArrayType            string `xml:"arrayType,attr"`
			ParameterValueStruct []struct {
				Name  string
				Value struct {
					Type  string `xml:"type,attr"`
					Value string `xml:",chardata"`
				}
			}
		}
	}
	err := xml.Unmarshal(b, &gpv)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decode GetParameterValuesResponse")
	}

	params := make([]datamodel.Parameter, 0, len(gpv.ParameterList.ParameterValueStruct))
	for _, pv := range gpv.ParameterList.ParameterValueStruct {
		params = append(params, datamodel.Parameter{
			Path:     pv.Name,
			Type:     pv.Value.Type,
			Value:    pv.Value.Value,
			Writable: true,
		})
	}

	isObject := func(name string, params []datamodel.Parameter) bool {
		prefix := name + "."
		for _, p := range params {
			if strings.HasPrefix(p.Path, prefix) {
				return true
			}
		}
		return false
	}
	exists := func(name string, params []datamodel.Parameter) bool {
		for _, p := range params {
			if p.Path == name {
				return true
			}
		}
		return false
	}
	for i, p := range params {
		params[i].Object = isObject(p.Path, params)
	}

	// Undefined objects
	for _, p := range params {
		tokens := strings.Split(p.Path, ".")
		if len(tokens) == 1 {
			if !exists(tokens[0], params) {
				params = append(params, datamodel.Parameter{
					Path:     tokens[0],
					Object:   true,
					Writable: true,
				})
			}
		} else {
			objPath := strings.Join(tokens[:len(tokens)-1], ".")
			if !exists(objPath, params) {
				params = append(params, datamodel.Parameter{
					Path:     objPath,
					Object:   true,
					Writable: true,
				})
			}
		}
	}

	return params
}
