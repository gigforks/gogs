package repo

import (
	"strings"

	api "github.com/gogits/go-gogs-client"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"
)

func DeleteCollaborator(ctx *context.APIContext, form api.AddCollaboratorOption) {
	collaborator, err := models.GetUserByName(ctx.Params(":collaborator"))
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.Error(422, "", err)
		} else {
			ctx.Error(500, "GetUserByName", err)
		}
		return
	}

	if err = ctx.Repo.Repository.DeleteCollaboration(collaborator.ID); err != nil {
		ctx.Status(204)
	}

}

// SearchOrgs returns Itsyou.Online organization names based on the scopes from the users
// access token
func SearchOrgs(ctx *context.APIContext) {
	q := ctx.Query("q")
	userOrgs := ctx.Session.Get("organizations").([]string)

	type OrgName struct {
		Name string
	}

	resp := make([]*OrgName, 0)
	for _, org := range userOrgs {
		if strings.HasPrefix(org, q) {
			resp = append(resp, &OrgName{Name: org})
		}
	}

	ctx.JSON(200, map[string]interface{}{
		"ok":   true,
		"data": resp,
	})
}
