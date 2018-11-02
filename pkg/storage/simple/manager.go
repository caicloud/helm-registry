/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package simple

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/lock"
	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/caicloud/helm-registry/pkg/storage"
	"github.com/caicloud/helm-registry/pkg/storage/driver"
	"k8s.io/helm/pkg/chartutil"

	"github.com/buger/jsonparser"
)

const managerName = "simple"
const chartPackageName = "chart.tgz"
const metadataName = "metadata.dat"
const valuesName = "values.dat"

// chart status
const statusName = ".status"
const (
	statusLocking = "LOCKING"
	statusSuccess = "SUCCESS"
)

// every resource has a named lock and the lock name specifies all spaces
const allSpacesLockName = ""

func init() {
	storage.Register(managerName, &simpleSpaceManagerFactory{})
}

// simpleSpaceManagerFactory implements storage.SpaceManagerFactory.
// Parameters has a specified parameter named storagedriver. Its value can be:
// inmemory, filesystem, s3, azure, swift, oss, gcs.
// For more information, please refer to [link](https://docs.docker.com/registry/storage-drivers/)
// All parameters will pass to docker registry StorageDriver for creating a storage backend.
// If you want to specify a backend driver, please find other parameters from above-mentioned link.
// For example, a simple manager uses filesystem storage driver as backend, parameters should be:
//  "storagedriver": "filesystem"
//  "rootdirectory": "/path/to/empty/dir"
//  "resourcelocker": "memory"
type simpleSpaceManagerFactory struct{}

// Create creates a new SpaceManager
func (factory *simpleSpaceManagerFactory) Create(parameters map[string]interface{}) (storage.SpaceManager, error) {
	if parameters == nil {
		return nil, ErrorNoParameter.Format("parameters")
	}
	// create resource locker
	lockerName, ok := parameters[common.ParameterResourceLocker]
	if !ok {
		return nil, ErrorContentMissing.Format(common.ParameterResourceLocker)
	}
	var lockerParams map[string]interface{}
	lockerParameters, ok := parameters[common.ParameterLockerParameters]
	if ok {
		lockerParams, _ = lockerParameters.(map[string]interface{})
	}
	locker, err := lock.Create(fmt.Sprint(lockerName), lockerParams)
	if err != nil {
		return nil, errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	// create storage driver
	storageDriverName, ok := parameters[common.ParameterNameStorageDriver]
	if !ok {
		return nil, ErrorContentMissing.Format(common.ParameterNameStorageDriver)
	}
	storageDriver, err := driver.Create(fmt.Sprint(storageDriverName), parameters)
	if err != nil {
		return nil, errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}

	var lockTimeout = lock.TimeoutImmediate
	paramLockTimeout, ok := parameters[common.ParameterLockTimeout]
	if ok {
		timeout, err := strconv.Atoi(fmt.Sprint(paramLockTimeout))
		if err != nil {
			return nil, errors.NewResponError(http.StatusBadRequest, "param.invalidate", "${name} invalidate", errors.M{
				"name": err,
			})
		}
		lockTimeout = time.Duration(timeout) * time.Millisecond
	}

	return NewSpaceManager(storageDriver, locker, lockTimeout), nil
}

// SpaceManager implements storage.SpaceManager interface, and stores charts in file system
type SpaceManager struct {
	Prefix      string
	Lock        lock.ResourceLocker
	LockTimeout time.Duration
	Backend     driver.StorageDriver
}

// NewSpaceManager creates a new SpaceManager
func NewSpaceManager(backend driver.StorageDriver, lock lock.ResourceLocker, timeout time.Duration) *SpaceManager {
	return &SpaceManager{"/", lock, timeout, backend}
}

// Kind returns kind name
func (sm *SpaceManager) Kind() string {
	return managerName
}

// Create creates a new Space with space name
func (sm *SpaceManager) Create(ctx context.Context, space string) (storage.Space, error) {
	lock := sm.Lock.Get(space)
	if !lock.Lock(sm.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "space.lock", "${name} lock", errors.M{
			"name": space,
		})
	}
	defer lock.Unlock()
	newSpace, err := sm.Space(ctx, space)
	if err != nil {
		return nil, err
	}
	if newSpace.Exists(ctx) {
		return nil, errors.NewResponError(http.StatusConflict, "charts.name.exists", "${name} exist", errors.M{
			"name": space,
		})
	}
	// space does not exist
	key := path.Join(sm.Prefix, space, statusName)
	err = sm.Backend.PutContent(ctx, key, []byte(statusSuccess))
	if err != nil {
		return nil, errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	return sm.Space(ctx, space)
}

// Delete deletes specific space.
func (sm *SpaceManager) Delete(ctx context.Context, space string) error {
	lock := sm.Lock.Get(space)
	if !lock.Lock(sm.LockTimeout) {
		return errors.NewResponError(http.StatusLocked, "space.lock", "${name} lock", errors.M{
			"name": space,
		})
	}
	defer lock.Unlock()
	return deleteKeys(ctx, sm.Backend, path.Join(sm.Prefix, space), true)
}

// List returns all space names
func (sm *SpaceManager) List(ctx context.Context) ([]string, error) {
	lock := sm.Lock.Get(allSpacesLockName)
	if !lock.RLock(sm.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "space.lock", "${name} lock", errors.M{
			"name": allSpacesLockName,
		})
	}
	defer lock.RUnlock()
	return list(ctx, sm.Backend, sm.Prefix, validateName, sortNames)
}

// Space returns a Space that it can manage specific space
func (sm *SpaceManager) Space(ctx context.Context, space string) (storage.Space, error) {
	if !validateName(space) {
		return nil, errors.NewResponError(http.StatusBadRequest, "space.invalidate", "${name} invalidate", errors.M{
			"name": space,
		})
	}
	return NewSpace(sm, space)
}

// Validate validates whether the value of vType is valid.
func (sm *SpaceManager) Validate(ctx context.Context, vType storage.ValidationType, value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	switch vType {
	case storage.ValidationTypeSpaceName, storage.ValidationTypeChartName:
		return validateName(str)
	case storage.ValidationTypeVersionNumber:
		return validateVersion(str)
	}
	return false
}

// Space defines methods for managing specific chart space
type Space struct {
	SpaceManager *SpaceManager
	Prefix       string
	Space        string
}

// NewSpace create a new Space
func NewSpace(spaceManager *SpaceManager, space string) (*Space, error) {
	if spaceManager == nil {
		return nil, ErrorNoParameter.Format("spaceManager")
	}
	if !validateName(space) {
		return nil, errors.NewResponError(http.StatusBadRequest, "space.invalidate", "${name} invalidate", errors.M{
			"name": space,
		})
	}
	return &Space{spaceManager, path.Join(spaceManager.Prefix, space), space}, nil
}

// Kind returns kind name
func (s *Space) Kind() string {
	return managerName
}

// Name returns name
func (s *Space) Name() string {
	return s.Space
}

// Delete deletes specific chart
func (s *Space) Delete(ctx context.Context, chart string) error {
	lock := s.SpaceManager.Lock.Get(s.Name(), chart)
	if !lock.Lock(s.SpaceManager.LockTimeout) {
		return errors.NewResponError(http.StatusLocked, "charts.lock", "${name} lock", errors.M{
			"name": s.Name() + "/" + chart,
		})
	}
	defer lock.Unlock()
	return deleteKeys(ctx, s.SpaceManager.Backend, path.Join(s.Prefix, chart), true)
}

// List returns all chart names
func (s *Space) List(ctx context.Context) ([]string, error) {
	lock := s.SpaceManager.Lock.Get(s.Name())
	if !lock.RLock(s.SpaceManager.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "space.lock", "${name} lock", errors.M{
			"name": s.Name(),
		})
	}
	defer lock.RUnlock()
	return list(ctx, s.SpaceManager.Backend, s.Prefix, validateName, sortNames)
}

// Exists returns whether the space exists
func (s *Space) Exists(ctx context.Context) bool {
	return keyExists(ctx, s.SpaceManager.Backend, s.Prefix)
}

// VersionMetadata returns all metadata of charts in current Space
func (s *Space) VersionMetadata(ctx context.Context) ([]*storage.Metadata, error) {
	list, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	mtAll := make([]*storage.Metadata, 0, len(list))
	for _, key := range list {
		chart, err := s.Chart(ctx, key)
		if err != nil {
			return nil, err
		}
		mtList, err := chart.VersionMetadata(ctx)
		if err != nil {
			return nil, err
		}
		mtAll = append(mtAll, mtList...)
	}
	return mtAll, nil
}

// Chart returns a Chart for managing specific chart
func (s *Space) Chart(ctx context.Context, chart string) (storage.Chart, error) {
	if !validateName(chart) {
		return nil, errors.NewResponError(http.StatusBadRequest, "charts.invalidate", "${name} invalidate", errors.M{
			"name": chart,
		})
	}
	return NewChart(s, chart)
}

// Chart defines methods for managing specific chart
type Chart struct {
	Space  *Space
	Prefix string
	Chart  string
}

// NewChart create a new Chart
func NewChart(space *Space, chart string) (*Chart, error) {
	if space == nil {
		return nil, ErrorNoParameter.Format("space")
	}
	if !validateName(chart) {
		return nil, errors.NewResponError(http.StatusBadRequest, "charts.invalidate", "${name} invalidate", errors.M{
			"name": chart,
		})
	}
	return &Chart{space, path.Join(space.Prefix, chart), chart}, nil
}

// Kind returns kind name
func (c *Chart) Kind() string {
	return managerName
}

// Name returns name
func (c *Chart) Name() string {
	return c.Chart
}

// Delete deletes specific chart
func (c *Chart) Delete(ctx context.Context, version string) error {
	lock := c.Space.SpaceManager.Lock.Get(c.Space.Name(), c.Name(), version)
	if !lock.Lock(c.Space.SpaceManager.LockTimeout) {
		return errors.NewResponError(http.StatusLocked, "version.lock", "${name} lock", errors.M{
			"name": c.Space.Name() + "/" + c.Name() + "/" + version,
		})
	}
	err := deleteKeys(ctx, c.Space.SpaceManager.Backend, path.Join(c.Prefix, version), true)
	// unlock before return
	lock.Unlock()
	if err != nil {
		return err
	}
	versions, err := c.List(ctx)
	if err == nil && len(versions) <= 0 {
		// delete chart if has no version
		return c.Space.Delete(ctx, c.Chart)
	}
	return err
}

// List returns all version numbers
func (c *Chart) List(ctx context.Context) ([]string, error) {
	lock := c.Space.SpaceManager.Lock.Get(c.Space.Name(), c.Name())
	if !lock.RLock(c.Space.SpaceManager.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "charts.lock", "${name} lock", errors.M{
			"name": c.Space.Name() + "/" + c.Name(),
		})
	}
	defer lock.RUnlock()
	return list(ctx, c.Space.SpaceManager.Backend, c.Prefix, validateVersion, sortVersions)
}

// Exists returns whether the chart exists
func (c *Chart) Exists(ctx context.Context) bool {
	return keyExists(ctx, c.Space.SpaceManager.Backend, c.Prefix)
}

// VersionMetadata returns all metadata of charts in current chart
func (c *Chart) VersionMetadata(ctx context.Context) ([]*storage.Metadata, error) {
	list, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	mtList := make([]*storage.Metadata, 0, len(list))
	for _, key := range list {
		version, err := c.Version(ctx, key)
		if err != nil {
			return nil, err
		}
		mt, err := version.Metadata(ctx)
		if err != nil {
			return nil, err
		}
		mtList = append(mtList, mt)
	}
	return mtList, nil
}

// Version returns a Version for managing specific version
func (c *Chart) Version(ctx context.Context, version string) (storage.Version, error) {
	if !validateVersion(version) {
		return nil, errors.NewResponError(http.StatusBadRequest, "version.invalidate", "${name} invalidate", errors.M{
			"name": version,
		})
	}
	return NewVersion(c, version)
}

// Version defines methods for managing specific version of a chart
type Version struct {
	Chart   *Chart
	Backend driver.StorageDriver
	Prefix  string
	Version string
}

// NewVersion creates new Version with chart and version name
func NewVersion(chart *Chart, version string) (*Version, error) {
	if chart == nil {
		return nil, ErrorNoParameter.Format("chart")
	}
	if !validateVersion(version) {
		return nil, errors.NewResponError(http.StatusBadRequest, "version.invalidate", "${name} invalidate", errors.M{
			"name": version,
		})
	}
	return &Version{chart, chart.Space.SpaceManager.Backend, path.Join(chart.Prefix, version), version}, nil
}

// Kind returns kind name
func (v *Version) Kind() string {
	return managerName
}

// Number returns version number
func (v *Version) Number() string {
	return v.Version
}

// PutContent stores chart data
func (v *Version) PutContent(ctx context.Context, data []byte) error {
	lock := v.Chart.Space.SpaceManager.Lock.Get(v.Chart.Space.Name(), v.Chart.Name(), v.Number())
	if !lock.Lock(v.Chart.Space.SpaceManager.LockTimeout) {
		return errors.NewResponError(http.StatusLocked, "version.lock", "${name} lock", errors.M{
			"name": v.Chart.Space.Name() + "/" + v.Chart.Name() + "/" + v.Number(),
		})
	}
	defer lock.Unlock()
	if len(data) <= 0 {
		return ErrorNoParameter.Format("data")
	}
	// Check whether process succeed
	var success = false
	defer func() {
		if !success {
			// GC when it's failed
			err := v.Chart.Delete(ctx, v.Version)
			if err != nil {
				log.Error(err)
			}
		}
	}()
	statusKey := path.Join(v.Prefix, statusName)
	statusData, err := v.Backend.GetContent(ctx, statusKey)
	if string(statusData) == statusLocking {
		return errors.NewResponError(http.StatusLocked, "charts.lock", "${name} lock", errors.M{
			"name": v.Chart.Name() + "/" + v.Version,
		})
	}
	// Create a `statusName` file with `statusLocking` to lock the place
	err = v.Backend.PutContent(ctx, statusKey, []byte(statusLocking))
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	// Validate chart
	chart, err := chartutil.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return errors.NewResponError(http.StatusBadRequest, "param.error", "${name} error", errors.M{
			"name": "chart",
		})
	}
	// Coalesce metadata
	metadata, err := storage.CoalesceMetadata(chart)
	if err != nil {
		return errors.NewResponError(http.StatusBadRequest, "charts.invalidate", "${name} error", errors.M{
			"name": chart,
		})
	}
	// Coalesce values
	values, err := chartutil.CoalesceValues(chart, chart.Values)
	if err != nil {
		return errors.NewResponError(http.StatusBadRequest, "param.invalidate", "${name} error", errors.M{
			"name": chart.Values,
		})
	}

	// Store chart
	err = v.Backend.PutContent(ctx, path.Join(v.Prefix, chartPackageName), data)
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	// Store metadata
	data, err = json.Marshal(metadata)
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	err = v.Backend.PutContent(ctx, path.Join(v.Prefix, metadataName), data)
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	// Store values
	data, err = json.Marshal(values)
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	err = v.Backend.PutContent(ctx, path.Join(v.Prefix, valuesName), data)
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	// Write `statusSuccess` to `statusName` file
	err = v.Backend.PutContent(ctx, statusKey, []byte(statusSuccess))
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	// Succeed in storing chart
	success = true
	return nil
}

// GetContent gets chart data
func (v *Version) GetContent(ctx context.Context) ([]byte, error) {
	lock := v.Chart.Space.SpaceManager.Lock.Get(v.Chart.Space.Name(), v.Chart.Name(), v.Number())
	if !lock.RLock(v.Chart.Space.SpaceManager.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "version.lock", "${name} lock", errors.M{
			"name": v.Chart.Space.Name() + "/" + v.Chart.Name() + "/" + v.Number(),
		})
	}
	defer lock.RUnlock()
	if err := v.Validate(ctx); err != nil {
		return nil, err
	}
	path := path.Join(v.Prefix, chartPackageName)
	data, err := v.Chart.Space.SpaceManager.Backend.GetContent(ctx, path)
	if err != nil {
		return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": v.Prefix,
		})
	}
	return data, nil
}

// Validate validates whether the chart is valid
func (v *Version) Validate(ctx context.Context) error {
	lock := v.Chart.Space.SpaceManager.Lock.Get(v.Chart.Space.Name(), v.Chart.Name(), v.Number())
	if !lock.RLock(v.Chart.Space.SpaceManager.LockTimeout) {
		return errors.NewResponError(http.StatusLocked, "version.lock", "${name} lock", errors.M{
			"name": v.Chart.Space.Name() + "/" + v.Chart.Name() + "/" + v.Number(),
		})
	}
	defer lock.RUnlock()
	data, err := v.Backend.GetContent(ctx, path.Join(v.Prefix, statusName))
	if err != nil {
		return errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": v.Prefix,
		})
	}
	status := string(data)
	if status != statusSuccess {
		return errors.NewResponError(http.StatusLocked, "charts.invalidate", "${name} lock", errors.M{
			"name": v.Chart.Space.Name() + "/" + v.Chart.Name() + "/" + v.Number(),
		})
	}
	return nil
}

// Exists returns whether the version exists
func (v *Version) Exists(ctx context.Context) bool {
	return keyExists(ctx, v.Backend, v.Prefix)
}

// Metadata returns a Metadata of current chart
func (v *Version) Metadata(ctx context.Context) (*storage.Metadata, error) {
	lock := v.Chart.Space.SpaceManager.Lock.Get(v.Chart.Space.Name(), v.Chart.Name(), v.Number())
	if !lock.RLock(v.Chart.Space.SpaceManager.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "version.lock", "${name} lock", errors.M{
			"name": v.Chart.Space.Name() + "/" + v.Chart.Name() + "/" + v.Number(),
		})
	}
	defer lock.RUnlock()
	if err := v.Validate(ctx); err != nil {
		return nil, err
	}

	pathValue := path.Join(v.Prefix, valuesName)
	dataValue, err := v.Backend.GetContent(ctx, pathValue)
	if err != nil {
		return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": err,
		})
	}

	typeStr, err := jsonparser.GetString(dataValue, "_config", "controllers", "[0]", "type")
	if err != nil {
		return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": err,
		})
	}

	path := path.Join(v.Prefix, metadataName)
	data, err := v.Backend.GetContent(ctx, path)
	if err != nil {
		return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": v.Prefix,
		})
	}
	meta := &storage.Metadata{}
	err = json.Unmarshal(data, meta)
	if err != nil {
		return nil, errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	meta.Type = typeStr

	return meta, nil
}

// Values gets data from values.yaml file which in current chart data
func (v *Version) Values(ctx context.Context) ([]byte, error) {
	lock := v.Chart.Space.SpaceManager.Lock.Get(v.Chart.Space.Name(), v.Chart.Name(), v.Number())
	if !lock.RLock(v.Chart.Space.SpaceManager.LockTimeout) {
		return nil, errors.NewResponError(http.StatusLocked, "version.lock", "${name} lock", errors.M{
			"name": v.Chart.Space.Name() + "/" + v.Chart.Name() + "/" + v.Number(),
		})
	}
	defer lock.RUnlock()
	if err := v.Validate(ctx); err != nil {
		return nil, err
	}
	path := path.Join(v.Prefix, valuesName)
	data, err := v.Backend.GetContent(ctx, path)
	if err != nil {
		return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": v.Prefix,
		})
	}
	return data, nil
}

var nameFilter = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// validateName validates whether the name can be used
func validateName(name string) bool {
	return nameFilter.MatchString(name)
}

var versionFilter = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// validateVersion validates whether the name can be used
func validateVersion(version string) bool {
	return versionFilter.MatchString(version)
}

// lastElement returns the last element of key. Its behavior like path.Base()
func lastElement(key string) string {
	key = strings.TrimRight(key, "/\\")
	index := strings.LastIndexAny(key, "/\\")
	if index >= 0 {
		key = key[index+1:]
	}
	return key
}

// list lists keys which only have one more element than prefix and return keys without prefix
func list(ctx context.Context, backend driver.StorageDriver, prefix string,
	validator func(string) bool, sorter func([]string) []string) ([]string, error) {
	list, err := backend.List(ctx, prefix)
	if err != nil {
		return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": prefix,
		})
	}
	i := 0
	for _, key := range list {
		key := lastElement(key)
		// filter invalid key
		if validator(key) {
			list[i] = key
			i++
		}
	}
	// valid keys
	list = list[:i]
	// sort keys
	if len(list) > 1 && sorter != nil {
		list = sorter(list)
	}
	return list, nil
}

// deleteKeys delete all keys by prefix if forced is true
func deleteKeys(ctx context.Context, backend driver.StorageDriver, prefix string, forced bool) error {
	list, err := backend.List(ctx, prefix)
	if err != nil {
		return errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
			"name": prefix,
		})
	}
	if len(list) > 0 && !forced {
		return ErrorNeedForcedDelete.Format(prefix)
	}
	err = backend.Delete(ctx, prefix)
	if err != nil {
		return errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": err.Error(),
		})
	}
	return nil
}

// keyExists check whether the key exists
func keyExists(ctx context.Context, backend driver.StorageDriver, key string) bool {
	_, err := backend.Stat(ctx, key)
	return err == nil
}

// StringSlice attaches the methods of Interface to []string, sorting in increasing order.
type StringSlice []string

func (p StringSlice) Len() int           { return len(p) }
func (p StringSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p StringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// parseVersionNumber parses version string to int
func parseVersionNumber(version string) []int {
	elements := strings.Split(version, ".")
	numbers := make([]int, 0, len(elements))
	for _, ele := range elements {
		num, err := strconv.Atoi(ele)
		if err != nil {
			// If came here, There is a bug in current manager.
			log.Panicln(err)
		}
		numbers = append(numbers, num)
	}
	return numbers
}

// VersionSlice attaches the methods of Interface to []string, sorting in increasing order.
type VersionSlice []string

func (p VersionSlice) Len() int { return len(p) }
func (p VersionSlice) Less(i, j int) bool {
	vni := parseVersionNumber(p[i])
	vnj := parseVersionNumber(p[j])
	for pos := 0; pos < len(vni) && pos < len(vnj); pos++ {
		switch {
		case vni[pos] < vnj[pos]:
			// less
			return true
		case vni[pos] > vnj[pos]:
			return false
		}
	}
	// equal
	return false
}
func (p VersionSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// sortNames sorts names by alphabetical order
func sortNames(slice []string) []string {
	sort.Sort(StringSlice(slice))
	return slice
}

// sortVersions sorts versions by version order
func sortVersions(slice []string) []string {
	sort.Sort(VersionSlice(slice))
	return slice
}
