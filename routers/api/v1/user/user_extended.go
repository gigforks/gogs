package user

import (
	api "github.com/gogits/go-gogs-client"
	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

// LIST ALL USERS FUNCTION
func ListAllUsers(ctx *context.APIContext) {
	lenUsers := models.CountUsers()
	// ALL IN ONE PAGE.

	if users, err := models.Users(0, int(lenUsers)); err == nil {
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

// PUT REQUEST TO ADD USER TO AN ORGANIZATION
func AddMyUserToOrganization(ctx *context.APIContext, form api.CreateOrgOption) {
	u := ctx.User
	orgname := form.UserName
	//fmt.Println("ADDMYUSERTOORG: Adding my user ", u.Name, " to org: ", orgname)
	if org, err := models.GetOrgByName(orgname); err == nil {
		if err = models.AddOrgUser(org.ID, u.ID); err == nil {
			ctx.JSON(200, u.APIFormat())
		} else {
			ctx.JSON(500, "Couldn't add user to org.")
		}
	}

}

// PUT REQUEST TO ADD USER TO AN ORGANIZATION
func AddUserToOrganization(ctx *context.APIContext, form api.CreateOrgOption) {
	username := ctx.Params(":username")
	orgname := form.UserName

	if u, err := models.GetUserByName(username); err == nil {
		if org, err := models.GetOrgByName(orgname); err == nil {
			if err = models.AddOrgUser(org.ID, u.ID); err == nil {
				ctx.JSON(200, u.APIFormat())
			} else {
				ctx.JSON(500, "Couldn't add user to org.")
			}
		}
	}

}

// DELETE REQUEST TO ADD CURRENT USER TO AN ORGANIZATION
func DeleteMyUserFromOrganization(ctx *context.APIContext, form api.CreateOrgOption) {
	u := ctx.User
	orgname := form.UserName
	if org, err := models.GetOrgByName(orgname); err == nil {
		if err = models.RemoveOrgUser(org.ID, u.ID); err == nil {
			ctx.JSON(200, u.APIFormat())
		} else {
			ctx.JSON(500, "Couldn't delete user from org.")
		}
	}

}

// DELETE REQUEST TO REMOVE USER FROM AN ORGANIZATION
func DeleteUserFromOrganization(ctx *context.APIContext, form api.CreateOrgOption) {
	username := ctx.Params(":username")
	orgname := form.UserName

	if u, err := models.GetUserByName(username); err == nil {
		if org, err := models.GetOrgByName(orgname); err == nil {
			if err = models.RemoveOrgUser(org.ID, u.ID); err == nil {
				ctx.JSON(200, u.APIFormat())
			} else {
				ctx.JSON(500, "Couldn't delete user from org.")

			}
		}
	}
}
