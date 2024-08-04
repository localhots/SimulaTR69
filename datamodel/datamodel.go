// Package datamodel supports CPE datamodel and state.
package datamodel

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/localhots/SimulaTR69/rpc"
)

type DataModel struct {
	Values       map[string]Parameter
	Bootstrapped bool

	version       Version
	commandKey    string
	events        []string
	notifyParams  []string
	retryAttempts uint32
	downUntil     time.Time
	lock          sync.RWMutex
}

type Version string

const (
	TR098 Version = "TR098"
	TR181 Version = "TR181"

	tr098Prefix = "InternetGatewayDevice."
	tr181Prefix = "Device."
)

//
// Accessors
//

func (dm *DataModel) Get(path string) []Parameter {
	dm.lock.RLock()
	defer dm.lock.RUnlock()

	params := []Parameter{}
	if strings.HasSuffix(path, ".") {
		for k, p := range dm.Values {
			if strings.HasPrefix(k, path) {
				params = append(params, p)
			}
		}
	} else if p, ok := dm.Values[path]; ok {
		params = append(params, p)
	}

	return params
}

func (dm *DataModel) GetValue(path string) Parameter {
	dm.lock.RLock()
	defer dm.lock.RUnlock()

	path = dm.prefixedPath(path)
	v, ok := dm.Values[path]
	if !ok {
		v = dm.newParameter(path)
	}
	return v
}

func (dm *DataModel) SetValue(path, val string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	path = dm.prefixedPath(path)
	v, ok := dm.Values[path]
	if !ok {
		v = dm.newParameter(path)
	}
	v.Value = val
	dm.Values[path] = v
}

func (dm *DataModel) SetValues(params []Parameter) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	for _, p := range params {
		v, ok := dm.Values[p.Path]
		if !ok {
			v = dm.newParameter(p.Path)
		}
		v.Type = p.Type
		v.Value = p.Value
		dm.Values[p.Path] = v
	}
}

func (dm *DataModel) CanSetValue(param Parameter) *rpc.FaultCode {
	dm.lock.RLock()
	defer dm.lock.RUnlock()

	v, ok := dm.Values[param.Path]
	if ok {
		if !v.Writable {
			return rpc.FaultNonWritableParameter.Ptr()
		}
		return nil
	}

	v, ok = dm.Values[dm.parent(param.Path)]
	if (ok && !v.Object) || !ok {
		return rpc.FaultInvalidParameterName.Ptr()
	}
	return nil
}

func (dm *DataModel) SetParameterAttribute(name string, notif int, notifChange bool, acl []string, aclChange bool) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if p, ok := dm.Values[name]; ok {
		if notifChange {
			p.Notification = notif
		}
		if aclChange {
			p.ACL = acl
		}
		dm.Values[name] = p
	}
}

func (dm *DataModel) AddObject(name string) (int, error) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	name = strings.TrimSuffix(name, ".")

	p, ok := dm.Values[name]
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
	var max int
	for k := range dm.Values {
		m := reg.FindStringSubmatch(k)
		if len(m) < 2 {
			continue
		}
		i, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if i > max {
			max = i
		}
	}

	next := max + 1
	newName := fmt.Sprintf("%s.%d", name, next)
	dm.Values[newName] = Parameter{
		Path:     newName,
		Object:   true,
		Writable: true,
	}

	return next, nil
}

func (dm *DataModel) DeleteObject(name string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	objName := strings.TrimSuffix(name, ".")
	for k := range dm.Values {
		// TODO: Improve this check. See if parent is writable
		if k == objName || strings.HasPrefix(k, name) {
			delete(dm.Values, k)
		}
	}
}

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

	dm.lock.RLock()
	defer dm.lock.RUnlock()

	params := []Parameter{}
	for k, p := range dm.Values {
		if reg.MatchString(k) {
			params = append(params, p)
		}
	}

	return params
}

//
// Events
//

func (dm *DataModel) AddEvent(evt string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if !slices.Contains(dm.events, evt) {
		dm.events = append(dm.events, evt)
	}
}

func (dm *DataModel) PendingEvents() []string {
	dm.lock.RLock()
	defer dm.lock.RUnlock()
	return dm.events
}

func (dm *DataModel) ClearEvents() {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.events = []string{}
}

//
// Bootstrap
//

func (dm *DataModel) IsBootstrapped() bool {
	return dm.Bootstrapped
}

func (dm *DataModel) SetBootstrapped(b bool) {
	dm.Bootstrapped = b
}

//
// Retry attempts
//

func (dm *DataModel) RetryAttempts() uint32 {
	return dm.retryAttempts
}

func (dm *DataModel) IncrRetryAttempts() {
	atomic.AddUint32(&dm.retryAttempts, 1)
}

func (dm *DataModel) ResetRetryAttempts() {
	atomic.SwapUint32(&dm.retryAttempts, 0)
}

//
// Command key
//

func (dm *DataModel) SetCommandKey(ck string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.commandKey = ck
}

func (dm *DataModel) GetCommandKey() string {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	return dm.commandKey
}

//
// Simulated downtime
//

func (dm *DataModel) SetDownUntil(du time.Time) {
	dm.downUntil = du
}

func (dm *DataModel) DownUntil() time.Time {
	return dm.downUntil
}

//
// Parameter change notification
//

func (dm *DataModel) NotifyParams() []string {
	return dm.notifyParams
}

func (dm *DataModel) NotifyParam(path string) {
	dm.notifyParams = append(dm.notifyParams, path)
}

func (dm *DataModel) ClearNotifyParams() {
	dm.notifyParams = []string{}
}

//
// Helpers
//

func (dm *DataModel) newParameter(path string) Parameter {
	return Parameter{
		Path:     path,
		Object:   false,
		Writable: true,
		Type:     rpc.TypeXSDString,
	}
}

func (dm *DataModel) detectVersion() {
	for k := range dm.Values {
		if strings.HasPrefix(k, tr098Prefix) {
			dm.version = TR098
			return
		}
		if strings.HasPrefix(k, tr181Prefix) {
			dm.version = TR181
			return
		}
	}
}

func (dm *DataModel) prefixedPath(path string) string {
	switch dm.version {
	case TR098:
		return tr098Prefix + path
	case TR181:
		return tr181Prefix + path
	default:
		return path
	}
}

func (dm *DataModel) firstValue(paths ...string) string {
	dm.lock.RLock()
	defer dm.lock.RUnlock()

	for _, path := range paths {
		if p, ok := dm.Values[path]; ok {
			return p.Value
		}
	}

	return ""
}

func (dm *DataModel) parent(path string) string {
	tokens := strings.Split(path, ".")
	return strings.Join(tokens[:len(tokens)-1], ".")
}

func (dm *DataModel) exists(path string) bool {
	_, ok := dm.Values[path]
	return ok
}
