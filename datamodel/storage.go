package datamodel

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

// Load looks or given datamodel and state paths and loads them.
func Load(dmPath, statePath string) (*DataModel, error) {
	log.Info().Str("file", dmPath).Msg("Loading datamodel")
	original, err := loadDataModel(dmPath)
	if err != nil {
		return nil, err
	}

	s, err := loadState(statePath)
	if err != nil {
		return nil, err
	}
	if s == nil {
		s = newState(original)
	} else {
		s.original = original
	}

	dm := &DataModel{values: s}
	dm.detectVersion()
	if !dm.IsBootstrapped() {
		dm.AddEvent(rpc.EventBootstrap)
	} else {
		dm.AddEvent(rpc.EventBoot)
	}

	return dm, nil
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

func loadState(filePath string) (*state, error) {
	if filePath == "" {
		return nil, nil
	}
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	b, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var s state
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}

	return &s, nil
}

func loadDataModel(filePath string) (map[string]Parameter, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("read datamodel file: %w", err)
	}
	defer fd.Close()
	r := csv.NewReader(fd)

	values := make(map[string]Parameter)
	var headerRead bool
	for {
		f, err := r.Read()
		// nolint:errorlint
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

		// Add parentPath object automatically if not defined explicitly
		parentPath := parent(p.Path)
		if _, ok := values[parentPath]; !ok {
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
