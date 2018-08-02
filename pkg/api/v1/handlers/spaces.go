/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"context"
	"path"
	"strings"

	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
)

// ListSpaces lists spaces
func ListSpaces(ctx context.Context) (int, []string, error) {
	return listStrings(ctx, func() ([]string, error) {
		spaces, err := common.MustGetSpaceManager().List(ctx)
		if err != nil {
			return nil, err
		}
		prefix := getTenantName(ctx) + "_"
		ret := make([]string, 0, 0)
		for i, space := range spaces {
			if space == SpecialTenantSpace {
				continue
			}
			if strings.HasPrefix(space, prefix) {
				_, name := splitSpace(spaces[i])
				ret = append(ret, name)
			}
		}
		return ret, nil
	})
}

// CreateSpace creates a specified space
func CreateSpace(ctx context.Context) (*models.Link, error) {
	name, err := getSpaceName(ctx)
	if err != nil {
		return nil, err
	}
	if tenant := getTenantName(ctx); tenant != SpecialTenant && name == SpecialTenantSpace {
		// Must not create a space named library when tenant is not system-tenant.
		return nil, errors.ErrorInvalidParam.Format("space", "no permission to create space "+SpecialSpace)
	}
	_, err = common.MustGetSpaceManager().Create(ctx, name)
	if err != nil {
		return nil, translateError(err, name)
	}
	link, err := getRequestPath(ctx)
	if err != nil {
		return nil, err
	}
	_, name = splitSpace(name)
	return models.NewLink(name, path.Join(link, name)), nil
}

// DeleteSpace deletes a specified space
func DeleteSpace(ctx context.Context) error {
	name, err := getSpaceName(ctx)
	if err != nil {
		return err
	}
	if tenant := getTenantName(ctx); tenant != SpecialTenant && name == SpecialTenantSpace {
		// Must not delete a space named library when tenant is not system-tenant.
		return errors.ErrorInvalidParam.Format("space", "no permission to delete space "+SpecialSpace)
	}
	return translateError(common.MustGetSpaceManager().Delete(ctx, name), name)
}
