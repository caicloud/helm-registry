/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package simple

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/caicloud/helm-registry/pkg/storage"
	"github.com/caicloud/helm-registry/pkg/storage/driver"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const managerName = "simple"
const chartPackageName = "chart.tgz"
const metadataName = "metadata.yaml"
const valuesName = "values.yaml"

// chart status
const statusName = ".status"
const (
	statusLocking = "LOCKING"
	statusSuccess = "SUCCESS"
)

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
type simpleSpaceManagerFactory struct{}

// Create creates a new SpaceManager
func (factory *simpleSpaceManagerFactory) Create(parameters map[string]interface{}) (storage.SpaceManager, error) {
	if parameters == nil {
		return nil, ErrorNoParameter.Format("parameters")
	}
	storageDriverName, ok := parameters["storagedriver"]
	if !ok {
		return nil, ErrorNoStorageDriver
	}
	backend := fmt.Sprint(storageDriverName)
	storageDriver, err := driver.Create(backend, parameters)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
	}
	return NewSpaceManager(storageDriver), nil
}

// SpaceManager implements storage.SpaceManager interface, and stores charts in file system
type SpaceManager struct {
	Prefix  string
	Backend driver.StorageDriver
}

// NewSpaceManager creates a new SpaceManager
func NewSpaceManager(backend driver.StorageDriver) *SpaceManager {
	return &SpaceManager{"/", backend}
}

// Name returns name of the manager
func (gm *SpaceManager) Name() string {
	return managerName
}

// Create creates a new Space with space name
func (gm *SpaceManager) Create(ctx context.Context, space string) (storage.Space, error) {
	if !validateName(space) {
		return nil, ErrorInvalidParam.Format("space", space)
	}
	key := path.Join(gm.Prefix, space, statusName)
	_, err := gm.Backend.Stat(ctx, key)
	if err == nil {
		return nil, ErrorResourceExist.Format(space)
	}
	// key does not exist
	err = gm.Backend.PutContent(ctx, key, []byte(statusSuccess))
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
	}
	return gm.Space(ctx, space)
}

// Delete deletes specific space.
func (gm *SpaceManager) Delete(ctx context.Context, space string) error {
	return deleteKeys(ctx, gm.Backend, path.Join(gm.Prefix, space), true)
}

// List returns all space names
func (gm *SpaceManager) List(ctx context.Context) ([]string, error) {
	return list(ctx, gm.Backend, gm.Prefix, validateName)
}

// Space returns a Space that it can manage specific space
func (gm *SpaceManager) Space(ctx context.Context, space string) (storage.Space, error) {
	if !validateName(space) {
		return nil, ErrorInvalidParam.Format("space", space)
	}
	return NewSpace(gm, space)
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
		return nil, ErrorInvalidParam.Format("space", space)
	}
	return &Space{spaceManager, path.Join(spaceManager.Prefix, space), space}, nil
}

// Name returns name
func (sm *Space) Name() string {
	return managerName
}

// Create creates a new Chart
func (sm *Space) Create(ctx context.Context, chart string) (storage.Chart, error) {
	return sm.Chart(ctx, chart)
}

// Delete deletes specific chart
func (sm *Space) Delete(ctx context.Context, chart string) error {
	return deleteKeys(ctx, sm.SpaceManager.Backend, path.Join(sm.Prefix, chart), true)
}

// List returns all chart names
func (sm *Space) List(ctx context.Context) ([]string, error) {
	return list(ctx, sm.SpaceManager.Backend, sm.Prefix, validateName)
}

// Charts returns all metadatas of charts in the current Space
func (sm *Space) Charts(ctx context.Context) ([]*chart.Metadata, error) {
	list, err := sm.List(ctx)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
	}
	mtAll := make([]*chart.Metadata, 0, len(list))
	for _, key := range list {
		cm, err := sm.Create(ctx, key)
		if err != nil {
			return nil, err
		}
		mtList, err := cm.Versions(ctx)
		if err != nil {
			return nil, err
		}
		mtAll = append(mtAll, mtList...)
	}
	return mtAll, nil
}

// Chart returns a Chart for managing specific chart
func (sm *Space) Chart(ctx context.Context, chart string) (storage.Chart, error) {
	if !validateName(chart) {
		return nil, ErrorInvalidParam.Format("chart", chart)
	}
	return NewChart(sm, chart)
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
		return nil, ErrorInvalidParam.Format("chart", chart)
	}
	return &Chart{space, path.Join(space.Prefix, chart), chart}, nil
}

// Name returns name
func (cm *Chart) Name() string {
	return managerName
}

// Create creates a new Version
func (cm *Chart) Create(ctx context.Context, version string) (storage.Version, error) {
	return cm.Version(ctx, version)
}

// Delete deletes specific chart
func (cm *Chart) Delete(ctx context.Context, version string) error {
	err := deleteKeys(ctx, cm.Space.SpaceManager.Backend, path.Join(cm.Prefix, version), true)
	if err != nil {
		return err
	}

	versions, err := cm.List(ctx)
	if err == nil && len(versions) <= 0 {
		// delete chart if has no version
		return cm.Space.Delete(ctx, cm.Chart)
	}
	return err
}

// List returns all version numbers
func (cm *Chart) List(ctx context.Context) ([]string, error) {
	return list(ctx, cm.Space.SpaceManager.Backend, cm.Prefix, validateVersion)
}

// Versions returns all metadatas of charts in the current chart
func (cm *Chart) Versions(ctx context.Context) ([]*chart.Metadata, error) {
	list, err := cm.List(ctx)
	if err != nil {
		return nil, err
	}
	mtList := make([]*chart.Metadata, 0, len(list))
	for _, key := range list {
		vm, err := cm.Create(ctx, key)
		if err != nil {
			return nil, err
		}
		mt, err := vm.Metadata(ctx)
		if err != nil {
			return nil, err
		}
		mtList = append(mtList, mt)
	}
	return mtList, nil
}

// Version returns a Version for managing specific version
func (cm *Chart) Version(ctx context.Context, version string) (storage.Version, error) {
	if !validateVersion(version) {
		return nil, ErrorInvalidParam.Format("version", version)
	}
	return NewVersion(cm, version)
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
		return nil, ErrorInvalidParam.Format("version")
	}
	return &Version{chart, chart.Space.SpaceManager.Backend, path.Join(chart.Prefix, version), version}, nil
}

// Name returns name
func (vm *Version) Name() string {
	return managerName
}

// PutContent stores chart data
func (vm *Version) PutContent(ctx context.Context, data []byte) error {
	if len(data) <= 0 {
		return ErrorNoParameter.Format("data")
	}
	// Check whether process succeed
	var success = false
	defer func() {
		if !success {
			// GC when it's failed
			err := vm.Chart.Delete(ctx, vm.Version)
			if err != nil {
				log.Error(err)
			}
		}
	}()
	statusKey := path.Join(vm.Prefix, statusName)
	statusData, err := vm.Backend.GetContent(ctx, statusKey)
	if err == nil && string(statusData) == statusLocking {
		return ErrorLocking.Format("chart", vm.Chart.Name()+"/"+vm.Version)
	}
	// Create a `statusName` file with `statusLocking` to lock the place
	err = vm.Backend.PutContent(ctx, statusKey, []byte(statusLocking))
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	// Extract Chat.yaml and values.yaml from data
	unzipped, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return ErrorParamTypeError.Format("chart", "gzip", "unknown")
	}
	defer unzipped.Close()
	reader := tar.NewReader(unzipped)
	var metadataBuffer, valuesBuffer *bytes.Buffer

	metadataPath := path.Join(vm.Chart.Chart, "Chart.yaml")
	valuesPath := path.Join(vm.Chart.Chart, "values.yaml")
	for {
		if hd, errn := reader.Next(); errn == nil {
			switch hd.Name {
			case metadataPath:
				metadataBuffer = bytes.NewBuffer(nil)
				if _, err = io.Copy(metadataBuffer, reader); err != nil {
					return ErrorInternalUnknown.Format(err)
				}
			case valuesPath:
				valuesBuffer = bytes.NewBuffer(nil)
				if _, err = io.Copy(valuesBuffer, reader); err != nil {
					return ErrorInternalUnknown.Format(err)
				}
			}
		} else {
			if errn == io.EOF {
				break
			}
			return ErrorInternalUnknown.Format(errn)
		}
		if metadataBuffer != nil && valuesBuffer != nil {
			break
		}
	}
	if metadataBuffer == nil {
		return ErrorNoResource.Format("Chart.yaml", vm.Prefix)
	}
	if valuesBuffer == nil {
		return ErrorNoResource.Format("values.yaml", vm.Prefix)
	}

	// Store chart
	err = vm.Backend.PutContent(ctx, path.Join(vm.Prefix, chartPackageName), data)
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	err = vm.Backend.PutContent(ctx, path.Join(vm.Prefix, metadataName), metadataBuffer.Bytes())
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	err = vm.Backend.PutContent(ctx, path.Join(vm.Prefix, valuesName), valuesBuffer.Bytes())
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	// Write `statusSuccess` to `statusName` file
	err = vm.Backend.PutContent(ctx, statusKey, []byte(statusSuccess))
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	// Succeed in storing chart
	success = true
	return nil
}

// GetContent gets chart data
func (vm *Version) GetContent(ctx context.Context) ([]byte, error) {
	if err := vm.Validate(ctx); err != nil {
		return nil, err
	}
	path := path.Join(vm.Prefix, chartPackageName)
	data, err := vm.Chart.Space.SpaceManager.Backend.GetContent(ctx, path)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
	}
	return data, nil
}

// Validate validates whether the chart is valid
func (vm *Version) Validate(ctx context.Context) error {
	data, err := vm.Backend.GetContent(ctx, path.Join(vm.Prefix, statusName))
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	status := string(data)
	if status != statusSuccess {
		return ErrorInvalidStatus.Format("chart", status)
	}
	return nil
}

// Metadata returns a Metadata of the current chart
func (vm *Version) Metadata(ctx context.Context) (*chart.Metadata, error) {
	if err := vm.Validate(ctx); err != nil {
		return nil, err
	}
	path := path.Join(vm.Prefix, metadataName)
	data, err := vm.Backend.GetContent(ctx, path)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
	}
	meta, err := chartutil.UnmarshalChartfile(data)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
	}
	return meta, nil
}

// Values gets data from values.yaml file which in the current chart data
func (vm *Version) Values(ctx context.Context) ([]byte, error) {
	if err := vm.Validate(ctx); err != nil {
		return nil, err
	}
	path := path.Join(vm.Prefix, valuesName)
	data, err := vm.Backend.GetContent(ctx, path)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
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

// lastElement returns the last element of key. Its behavior like path.Bash()
func lastElement(key string) string {
	key = strings.TrimRight(key, "/\\")
	index := strings.LastIndexAny(key, "/\\")
	if index >= 0 {
		key = key[index+1:]
	}
	return key
}

// list lists keys which only have one more element than prefix and return keys without prefix
func list(ctx context.Context, backend driver.StorageDriver, prefix string, validator func(string) bool) ([]string, error) {
	list, err := backend.List(ctx, prefix)
	if err != nil {
		return nil, ErrorInternalUnknown.Format(err)
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
	return list[:i], nil
}

// deleteKeys delete all keys by prefix if forced is true
func deleteKeys(ctx context.Context, backend driver.StorageDriver, prefix string, forced bool) error {
	list, err := backend.List(ctx, prefix)
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	if len(list) > 0 && !forced {
		return ErrorNeedForcedDelete.Format(prefix)
	}
	err = backend.Delete(ctx, prefix)
	if err != nil {
		return ErrorInternalUnknown.Format(err)
	}
	return nil
}
