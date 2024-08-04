package datamodel

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/localhots/SimulaTR69/rpc"
	"github.com/rs/zerolog/log"
)

type DataModel struct {
	Version       DataModelVersion
	Bootstrapped  bool
	RetryAttempts uint32
	CommandKey    string
	Events        []string
	NotifyParams  []string
	Values        map[string]Parameter
	DownUntil     time.Time

	lock sync.RWMutex
}

type DeviceID struct {
	Manufacturer string
	OUI          string
	ProductClass string
	SerialNumber string
}

type Parameter struct {
	Path         string
	Object       bool
	Writable     bool
	Type         string
	Value        string
	Notification int
	ACL          []string
}

type DataModelVersion string

const (
	TR098 DataModelVersion = "TR098"
	TR181 DataModelVersion = "TR181"

	tr098Prefix = "InternetGatewayDevice."
	tr181Prefix = "Device."
)

func LoadDataModel(dmPath, statePath string) (*DataModel, error) {
	log.Info().Str("file", dmPath).Msg("Loading datamodel")
	dm, err := loadState(statePath)
	if err != nil {
		return nil, err
	}
	if dm == nil {
		dm, err = loadDataModel(dmPath)
		if err != nil {
			return nil, err
		}
	}

	dm.detectVersion()
	if !dm.Bootstrapped {
		dm.AddEvent(rpc.EventBootstrap)
	} else {
		dm.AddEvent(rpc.EventBoot)
	}

	return dm, nil
}

func loadState(filePath string) (*DataModel, error) {
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

	var dm DataModel
	if err := json.Unmarshal(b, &dm); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}

	return &dm, nil
}

func loadDataModel(filePath string) (*DataModel, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("read datamodel file: %w", err)
	}
	defer fd.Close()
	r := csv.NewReader(fd)

	dm := DataModel{Values: make(map[string]Parameter)}
	var headerRead bool
	for {
		f, err := r.Read()
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
		dm.Values[p.Path] = p
	}

	return &dm, nil
}

func (dm *DataModel) SaveState(stateFile string) error {
	if stateFile == "" {
		return nil
	}

	b, err := json.MarshalIndent(dm, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal datamodel: %w", err)
	}

	if err := os.WriteFile(stateFile, b, 0600); err != nil {
		return fmt.Errorf("save state file: %w", err)
	}
	return nil
}

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

func (dm *DataModel) TrySetValue(param Parameter) *rpc.FaultCode {
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

func (dm *DataModel) DeviceID() DeviceID {
	return DeviceID{
		Manufacturer: dm.firstValue(
			"DeviceID.Manufacturer",
			"Device.DeviceInfo.Manufacturer",
			"InternetGatewayDevice.DeviceInfo.Manufacturer",
		),
		OUI: dm.firstValue(
			"DeviceID.OUI",
			"Device.DeviceInfo.ManufacturerOUI",
			"InternetGatewayDevice.DeviceInfo.ManufacturerOUI",
		),
		ProductClass: dm.firstValue(
			"DeviceID.ProductClass",
			"Device.DeviceInfo.ProductClass",
			"InternetGatewayDevice.DeviceInfo.ProductClass",
		),
		SerialNumber: dm.firstValue(
			"DeviceID.SerialNumber",
			"Device.DeviceInfo.SerialNumber",
			"InternetGatewayDevice.DeviceInfo.SerialNumber",
		),
	}
}

func (dm *DataModel) AddEvent(evt string) {
	dm.lock.Lock()
	defer dm.lock.Unlock()

	if !slices.Contains(dm.Events, evt) {
		dm.Events = append(dm.Events, evt)
	}
}

func (dm *DataModel) PendingEvents() []string {
	dm.lock.RLock()
	defer dm.lock.RUnlock()
	return dm.Events
}

func (dm *DataModel) ClearEvents() {
	dm.lock.Lock()
	defer dm.lock.Unlock()
	dm.Events = []string{}
}

func (dm *DataModel) IncrRetryAttempts() {
	atomic.AddUint32(&dm.RetryAttempts, 1)
}

func (dm *DataModel) ResetRetryAttempts() {
	atomic.SwapUint32(&dm.RetryAttempts, 0)
}

func (dm *DataModel) SetSerialNumber(val string) {
	dm.SetValue("DeviceInfo.SerialNumber", val)
}

func (dm *DataModel) ConnectionRequestURL() Parameter {
	return dm.GetValue("ManagementServer.ConnectionRequestURL")
}

func (dm *DataModel) SetConnectionRequestURL(val string) {
	dm.SetValue("ManagementServer.ConnectionRequestURL", val)
}

func (dm *DataModel) SetParameterKey(val string) {
	dm.SetValue("ManagementServer.ParameterKey", val)
}

func (dm *DataModel) PeriodicInformEnabled() bool {
	val := dm.GetValue("ManagementServer.PeriodicInformEnable")
	b, _ := strconv.ParseBool(val.Value)
	return b
}

func (dm *DataModel) PeriodicInformInterval() time.Duration {
	const defaultInterval = 5 * time.Minute
	const secondsInDay = int64(24 * time.Hour / time.Second)
	val := dm.GetValue("ManagementServer.PeriodicInformInterval")
	i, _ := strconv.ParseInt(val.Value, 10, 32)
	if i == 0 || i > secondsInDay {
		return defaultInterval
	}
	return time.Duration(i) * time.Second
}

func (dm *DataModel) SetPeriodicInformInterval(sec int64) {
	dm.SetValue("ManagementServer.PeriodicInformInterval", strconv.FormatInt(sec, 10))
}

func (dm *DataModel) PeriodicInformTime() time.Time {
	val := dm.GetValue("ManagementServer.PeriodicInformTime")
	i, _ := strconv.ParseInt(val.Value, 10, 32)
	return time.Unix(i, 0)
}

func (dm *DataModel) SetPeriodicInformTime(ts time.Time) {
	dm.SetValue("ManagementServer.PeriodicInformTime", strconv.FormatInt(ts.Unix(), 10))
}

func (dm *DataModel) IsPeriodicInformParameter(name string) bool {
	if strings.HasSuffix(name, "ManagementServer.PeriodicInformInterval") {
		return true
	}
	if strings.HasSuffix(name, "ManagementServer.PeriodicInformTime") {
		return true
	}
	if strings.HasSuffix(name, "ManagementServer.PeriodicInformEnable") {
		return true
	}
	return false
}

func (dm *DataModel) SetFirmwareVersion(ver string) {
	dm.SetValue("DeviceInfo.SoftwareVersion", ver)
}

func (dm *DataModel) GetValue(path string) Parameter {
	path = dm.prefixedPath(path)
	v, ok := dm.Values[path]
	if !ok {
		v = dm.newParameter(path)
	}
	return v
}

func (dm *DataModel) SetValue(path, val string) {
	path = dm.prefixedPath(path)
	v, ok := dm.Values[path]
	if !ok {
		v = dm.newParameter(path)
	}
	v.Value = val
	dm.Values[path] = v
}

func (dm *DataModel) TrimPrefix(path string) string {
	path = strings.TrimPrefix(path, tr098Prefix)
	path = strings.TrimPrefix(path, tr181Prefix)
	return path
}

func (dm *DataModel) TrimPrefixes(paths []string) []string {
	trimmed := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimPrefix(path, tr098Prefix)
		path = strings.TrimPrefix(path, tr181Prefix)
		trimmed = append(trimmed, path)
	}
	return trimmed
}

func (dm *DataModel) SetCommandKey(ck string) {
	dm.CommandKey = ck
}

func (dm *DataModel) GetCommandKey() string {
	return dm.CommandKey
}

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
			dm.Version = TR098
			return
		}
		if strings.HasPrefix(k, tr181Prefix) {
			dm.Version = TR181
			return
		}
	}
}

func (dm *DataModel) prefixedPath(path string) string {
	switch dm.Version {
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

func (p Parameter) Encode() rpc.ParameterValueEncoder {
	return rpc.ParameterValueEncoder{
		Name: p.Path,
		Value: rpc.ValueEncoder{
			Type:  p.Type,
			Value: p.Value,
		},
	}
}
