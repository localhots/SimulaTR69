package datamodel

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/localhots/blip/noctx/log"
)

// LoadState loads the state from the specified file path. If the file path
// is empty or the file does not exist, it returns a new state. If there is
// an error reading or parsing the file, it returns an error.
func LoadState(filePath string) (*State, error) {
	if filePath == "" {
		return newState(), nil
	}
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return newState(), nil
	}

	// Assume the file is trusted
	//nolint:gosec
	b, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}

	return &s, nil
}

// LoadDataModelFile loads the data model from the specified file path.
func LoadDataModelFile(filePath string) (map[string]Parameter, error) {
	// Assume the file is trusted
	//nolint:gosec
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("read datamodel file: %w", err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			log.Error("Failed to close datamodel file", log.Cause(err), log.F{"path": filePath})
		}
	}()

	return LoadDataModel(fd)
}

// LoadDataModel reads the data model from the provided io.Reader and returns
// a map of parameters. It expects the data to be in CSV format with a header
// row. Each row should contain the path, object flag, writable flag, value,
// and type of the parameter. If there is an error reading or parsing the CSV
// data, it returns an error.
func LoadDataModel(r io.Reader) (map[string]Parameter, error) {
	csvr := csv.NewReader(r)

	values := make(map[string]Parameter)
	var headerRead bool
	for {
		f, err := csvr.Read()
		//nolint:errorlint
		if err == io.EOF {
			break
		}
		if !headerRead {
			headerRead = true
			continue
		}

		isObject, err := strconv.ParseBool(f[1])
		if err != nil {
			return nil, fmt.Errorf("parse bool %q: %w", f[1], err)
		}
		writable, err := strconv.ParseBool(f[2])
		if err != nil {
			return nil, fmt.Errorf("parse bool %q: %w", f[2], err)
		}
		p := Parameter{
			Path:     f[0],
			Object:   isObject,
			Writable: writable,
			Type:     f[4],
			Value:    f[3],
		}
		if err := p.initGenerator(); err != nil {
			return nil, fmt.Errorf("init generator: %w", err)
		}

		// Add parentPath object automatically if not defined explicitly
		parentPath := parent(p.Path)
		if _, ok := values[parentPath]; !ok && parentPath != "" {
			values[parentPath] = Parameter{
				Path:     parentPath,
				Object:   true,
				Writable: true,
			}
		}

		values[p.Path] = p
	}

	return values, nil
}

// SaveState saves the state to the given file.
func (dm *DataModel) SaveState(stateFile string) error {
	if stateFile == "" {
		return nil
	}

	b, err := json.MarshalIndent(dm.values, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal datamodel: %w", err)
	}

	if err := os.WriteFile(stateFile, b, 0600); err != nil {
		return fmt.Errorf("save state file: %w", err)
	}
	return nil
}
