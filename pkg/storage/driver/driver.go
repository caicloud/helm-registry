/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package driver

import (
	"io"

	"github.com/docker/distribution/context"
	storageDriver "github.com/docker/distribution/registry/storage/driver"
)

// StorageDriver defines methods that a Storage Driver must implement for a
// filesystem-like key/value object storage.
// The interface is a copy of docker-registry StorageDriver. It makes docker
// drivers appear to be part of helm-registry
type StorageDriver interface {
	// Name returns the human-readable "name" of the driver, useful in error
	// messages and logging. By convention, this will just be the registration
	// name, but drivers may provide other information here.
	Name() string

	// GetContent retrieves the content stored at "path" as a []byte.
	// This should primarily be used for small objects.
	GetContent(ctx context.Context, path string) ([]byte, error)

	// PutContent stores the []byte content at a location designated by "path".
	// This should primarily be used for small objects.
	PutContent(ctx context.Context, path string, content []byte) error

	// Reader retrieves an io.ReadCloser for the content stored at "path"
	// with a given byte offset.
	// May be used to resume reading a stream by providing a nonzero offset.
	Reader(ctx context.Context, path string, offset int64) (io.ReadCloser, error)

	// Writer returns a FileWriter which will store the content written to it
	// at the location designated by "path" after the call to Commit.
	Writer(ctx context.Context, path string, append bool) (storageDriver.FileWriter, error)

	// Stat retrieves the FileInfo for the given path, including the current
	// size in bytes and the creation time.
	Stat(ctx context.Context, path string) (storageDriver.FileInfo, error)

	// List returns a list of the objects that are direct descendants of the
	//given path.
	List(ctx context.Context, path string) ([]string, error)

	// Move moves an object stored at sourcePath to destPath, removing the
	// original object.
	// Note: This may be no more efficient than a copy followed by a delete for
	// many implementations.
	Move(ctx context.Context, sourcePath string, destPath string) error

	// Delete recursively deletes all objects stored at "path" and its subpaths.
	Delete(ctx context.Context, path string) error

	// URLFor returns a URL which may be used to retrieve the content stored at
	// the given path, possibly using the given options.
	// May return an ErrUnsupportedMethod in certain StorageDriver
	// implementations.
	URLFor(ctx context.Context, path string, options map[string]interface{}) (string, error)
}
