package repo

import (
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
