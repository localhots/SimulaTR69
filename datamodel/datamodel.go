package datamodel

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/localhots/SimulaTR69/rpc"
)

// DataModel describes a stateful CPE datamodel.
type DataModel struct {
	values        *State
	version       version
	commandKey    string
	events        []string
	retryAttempts uint32
	downUntil     time.Time
	lock          sync.RWMutex
}

// version is a datamodel version identifier.
type version string

const (
	unknownVersion version = "Unknown"
	tr098          version = "TR-098"
	tr181          version = "TR-181"

	tr098Prefix = "InternetGatewayDevice."
	tr181Prefix = "Device."
)

//
// Accessors
//

// New creates a new DataModel instance with the given state.
func New(state *State) *DataModel {
	dm := &DataModel{values: state}
	dm.init()
	return dm
}

// Reset clears the DataModel to its initial state.
func (dm *DataModel) Reset() {
	dm.values.reset()
	dm.version = unknownVersion
	dm.commandKey = ""
	dm.events = []string{}
	dm.retryAttempts = 0
	dm.downUntil = time.Time{}
	dm.init()
}

// Version returns the current version of the DataModel.
func (dm *DataModel) Version() string {
	return string(dm.version)
}

// GetAll returns one or more parameters prefixed with the given path.
func (dm *DataModel) GetAll(path string) (params []Parameter, ok bool) {
	if !strings.HasSuffix(path, ".") {
		p, ok := dm.values.get(path)
		return []Parameter{p}, ok
	}

	dm.values.forEach(func(p Parameter) (cont bool) {
		if strings.HasPrefix(p.Path, path) {
			params = append(params, p)
		}
		return true
	})

	return params, len(params) > 0
}

// GetValue returns a parameter value with the given path and a boolean that
// is equal to true if a parameter exists.
func (dm *DataModel) GetValue(path string) (p Parameter, ok bool) {
	return dm.values.get(dm.prefixedPath(path))
}

// GetValues returns all parameters in the given paths. If at least one
// requested parameter is missing ok will be set to false.
func (dm *DataModel) GetValues(paths ...string) (params []Parameter, ok bool) {
	res := make(map[string]Parameter)
	allOK := true
	for _, path := range paths {
		if p, ok := dm.GetValue(path); ok {
			res[path] = p
		} else {
			allOK = false
		}
	}

	return slices.Collect(maps.Values(res)), allOK
}

// SetValue sets the value of a given parameter.
func (dm *DataModel) SetValue(path, val string) {
	path = dm.prefixedPath(path)
	param, ok := dm.values.get(path)
	if !ok {
		param = newParameter(path)
	}
	param.Value = val
	dm.values.save(param)
}

// SetValues saves multiple parameter values.
func (dm *DataModel) SetValues(params []Parameter) {
	for _, p := range params {
		v, ok := dm.values.get(p.Path)
		if !ok {
			v = newParameter(p.Path)
		}
		v.Type = p.Type
		v.Value = p.Value
		dm.values.save(v)
	}
}

// CanSetValue returns a non-nil fault code if a value can't be set.
func (dm *DataModel) CanSetValue(param Parameter) *rpc.FaultCode {
	v, ok := dm.values.get(param.Path)
	if ok {
		if !v.Writable {
			return rpc.FaultNonWritableParameter.Ptr()
		}
		return nil
	}

	v, ok = dm.values.get(parent(param.Path))
	if (ok && !v.Object) || !ok {
		return rpc.FaultInvalidParameterName.Ptr()
	}
	return nil
}

// SetParameterAttribute changes parameter value attributes.
func (dm *DataModel) SetParameterAttribute(name string, notif int, notifChange bool, acl []string, aclChange bool) {
	if p, ok := dm.values.get(name); ok {
		if notifChange {
			p.Notification = rpc.AttributeNotification(notif)
		}
		if aclChange {
			p.ACL = acl
		}
		dm.values.save(p)
	}
}

// AddObject create a new object and returns the index if successful.
func (dm *DataModel) AddObject(name string) (int, error) {
	name = strings.TrimSuffix(name, ".")

	p, ok := dm.values.get(name)
	if !ok {
		return 0, errors.New("parent object doesn't exist")
	}
	if !p.Object {
		return 0, errors.New("parent is not an object")
	}
	if !p.Writable {
		return 0, errors.New("parent is not writable")
	}

	reg := regexp.MustCompile(`^` + name + `\.(\d+)`)
	var maxIndex int
	dm.values.forEach(func(p Parameter) (cont bool) {
		m := reg.FindStringSubmatch(p.Path)
		if len(m) < 2 {
			return true
		}
		i, err := strconv.Atoi(m[1])
		if err != nil {
			return true
		}
		if i > maxIndex {
			maxIndex = i
		}
		return true
	})

	next := maxIndex + 1
	newName := fmt.Sprintf("%s.%d", name, next)
	dm.values.save(Parameter{
		Path:     newName,
		Object:   true,
		Writable: true,
	})

	return next, nil
}

// DeleteObject deletes the given object.
func (dm *DataModel) DeleteObject(name string) {
	objName := strings.TrimSuffix(name, ".")
	// TODO: Improve this check. See if parent is writable
	dm.values.delete(objName)
	dm.values.deletePrefix(name)
}

// ParameterNames returns all subparameters in the given path. If nextLevel is
// set to true the list of parameters goes one level deeper.
// nolint:nestif
func (dm *DataModel) ParameterNames(path string, nextLevel bool) []Parameter {
	var reg *regexp.Regexp
	if path == "" {
		if nextLevel {
			reg = regexp.MustCompile(`^[^\.]+$`)
		} else {
			reg = regexp.MustCompile(`.*`)
		}
	} else {
		path = strings.TrimSuffix(path, ".")
		path = strings.ReplaceAll(path, ".", "\\.")

		if nextLevel {
			reg = regexp.MustCompile(`^` + path + `\.[^\.]+$`)
		} else {
			reg = regexp.MustCompile(`^` + path + `\..*`)
		}
	}

	params := []Parameter{}
	dm.values.forEach(func(p Parameter) (cont bool) {
		if reg.MatchString(p.Path) {
			params = append(params, p)
		}
		return true
	})
	return params
}

//
// Events
//

// PendingEvents returns all events to be advertised during the next inform
// message.
func (dm *DataModel) PendingEvents() []string {
	dm.lock.RLock()
	defer dm.lock.RUnlock()
	return dm.events
}

// AddEvent adds a new event to be advertised during the next inform message.
func (dm *DataModel) AddEvent(evt string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if !slices.Contains(dm.events, evt) {
		dm.events = append(dm.events, evt)
	}
}

// ClearEvents removes all pending inform events.
func (dm *DataModel) ClearEvents() {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.events = []string{}
}

//
// Bootstrap
//

// IsBootstrapped returns true if CPE is had a successful bootstrap message
// exchange.
func (dm *DataModel) IsBootstrapped() bool {
	return dm.values.Bootstrapped
}

// SetBootstrapped assigns the bootstrap flag to the given value.
func (dm *DataModel) SetBootstrapped(b bool) {
	dm.values.Bootstrapped = b
}

//
// Retry attempts
//

// RetryAttempts returns the number of currently take attepts to inform.
func (dm *DataModel) RetryAttempts() uint32 {
	return dm.retryAttempts
}

// IncrRetryAttempts increments the number of infrom attempts by one.
func (dm *DataModel) IncrRetryAttempts() {
	atomic.AddUint32(&dm.retryAttempts, 1)
}

// ResetRetryAttempts resets the number of infrom attempts to zero.
func (dm *DataModel) ResetRetryAttempts() {
	atomic.SwapUint32(&dm.retryAttempts, 0)
}

//
// Command key
//

// CommandKey returns the current command key.
func (dm *DataModel) CommandKey() string {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	return dm.commandKey
}

// SetCommandKey sets the command key value.
func (dm *DataModel) SetCommandKey(ck string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.commandKey = ck
}

//
// Simulated downtime
//

// DownUntil returns the time the CPE will stop pretending to be offline.
func (dm *DataModel) DownUntil() time.Time {
	return dm.downUntil
}

// SetDownUntil sets the time until the CPE should pretend to be offline.
func (dm *DataModel) SetDownUntil(du time.Time) {
	dm.downUntil = du
}

//
// Helpers
//

func (dm *DataModel) init() {
	dm.detectVersion()
}

func (dm *DataModel) detectVersion() {
	dm.values.forEach(func(p Parameter) (cont bool) {
		if strings.HasPrefix(p.Path, tr098Prefix) {
			dm.version = tr098
			return false
		}
		if strings.HasPrefix(p.Path, tr181Prefix) {
			dm.version = tr181
			return false
		}
		return true
	})
	if dm.version == "" {
		dm.version = unknownVersion
	}
}

func (dm *DataModel) prefixedPath(path string) string {
	switch dm.version {
	case tr098:
		if strings.HasPrefix(path, tr098Prefix) {
			return path
		}
		return tr098Prefix + path
	case tr181:
		if strings.HasPrefix(path, tr181Prefix) {
			return path
		}
		return tr181Prefix + path
	default:
		return path
	}
}

func (dm *DataModel) firstValue(paths ...string) string {
	for _, path := range paths {
		if p, ok := dm.values.get(path); ok {
			return p.GetValue()
		}
	}

	return ""
}

func newParameter(path string) Parameter {
	return Parameter{
		Path:     path,
		Object:   false,
		Writable: true,
		Type:     rpc.XSD(rpc.TypeString),
	}
}

func parent(path string) string {
	tokens := strings.Split(path, ".")
	return strings.Join(tokens[:len(tokens)-1], ".")
}
