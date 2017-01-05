package org

import (
	"fmt"

	api "github.com/gogits/go-gogs-client"
	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/routers/api/v1/convert"
)

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
