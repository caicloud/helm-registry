/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"context"
	"path"

	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/common"
)

// ListSpaces lists spaces
func ListSpaces(ctx context.Context) (int, []string, error) {
	return listStrings(ctx, "spaces", func() ([]string, error) {
		return common.MustGetSpaceManager().List(ctx)
	})
}

// CreateSpace creates a specified space
func CreateSpace(ctx context.Context) (*models.Link, error) {
	name, err := getSpaceName(ctx)
	if err != nil {
		return nil, err
	}
	_, err = common.MustGetSpaceManager().Create(ctx, name)
	if err != nil {
		return nil, err
	}
	link, err := getRequestPath(ctx)
	if err != nil {
		return nil, err
	}
	return models.NewLink(name, path.Join(link, name)), nil
}

// DeleteSpace deletes a specified space
func DeleteSpace(ctx context.Context) error {
	name, err := getSpaceName(ctx)
	if err != nil {
		return err
	}
	return common.MustGetSpaceManager().Delete(ctx, name)
}
