// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package org

import (
	"fmt"

	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/routers/api/v1/convert"
	"github.com/gogits/gogs/routers/api/v1/user"
)

func listUserOrgs(ctx *context.APIContext, u *models.User, all bool) {
	if err := u.GetOrganizations(all); err != nil {
		ctx.Error(500, "GetOrganizations", err)
		return
	}

	apiOrgs := make([]*api.Organization, len(u.Orgs))
	for i := range u.Orgs {
		apiOrgs[i] = convert.ToOrganization(u.Orgs[i])
	}
	ctx.JSON(200, &apiOrgs)
}

// https://github.com/gogits/go-gogs-client/wiki/Organizations#list-your-organizations
func ListMyOrgs(ctx *context.APIContext) {
	listUserOrgs(ctx, ctx.User, true)
}

// https://github.com/gogits/go-gogs-client/wiki/Organizations#list-user-organizations
func ListUserOrgs(ctx *context.APIContext) {
	u := user.GetUserByParams(ctx)
	if ctx.Written() {
		return
	}
	listUserOrgs(ctx, u, false)
}

func ListAllOrgs(ctx *context.APIContext) {
	lenOrgs := int(models.CountOrganizations())

	orgSearchOpts := &models.SearchUserOptions{
		Keyword:  "_",
		Type:     models.USER_TYPE_ORGANIZATION,
		PageSize: lenOrgs,
		Page:     0,
	}
	fmt.Println(orgSearchOpts)
	// ALL IN ONE PAGE.
	if users, _, err := models.SearchUserByName(orgSearchOpts); err == nil {
		results := make([]*api.User, len(users))
		for i := range users {
			results[i] = &api.User{
				ID:        users[i].ID,
				UserName:  users[i].Name,
				AvatarUrl: users[i].AvatarLink(),
				FullName:  users[i].FullName,
			}
			if ctx.IsSigned {
				results[i].Email = users[i].Email
			}
		}
		ctx.JSON(200, results)
	}

}

// https://github.com/gogits/go-gogs-client/wiki/Organizations#get-an-organization
func Get(ctx *context.APIContext) {
	ctx.JSON(200, convert.ToOrganization(ctx.Org.Organization))
}

func CreateOrganization(ctx *context.APIContext, form api.CreateOrgOption) {
	orgInfo := &models.User{
		Name:        form.UserName,
		Description: form.Description,
		IsActive:    true,
		Type:        models.USER_TYPE_ORGANIZATION,
	}
	currentUser := ctx.User
	if err := models.CreateOrganization(orgInfo, currentUser); err != nil {
		ctx.JSON(500, "Couldn't create organization.")
	}
	ctx.JSON(200, convert.ToOrganization(orgInfo))
}

// https://github.com/gogits/go-gogs-client/wiki/Organizations#edit-an-organization
func Edit(ctx *context.APIContext, form api.EditOrgOption) {
	org := ctx.Org.Organization
	if !org.IsOwnedBy(ctx.User.ID) {
		ctx.Status(403)
		return
	}

	org.FullName = form.FullName
	org.Description = form.Description
	org.Website = form.Website
	org.Location = form.Location
	if err := models.UpdateUser(org); err != nil {
		ctx.Error(500, "UpdateUser", err)
		return
	}

	ctx.JSON(200, convert.ToOrganization(org))
}
