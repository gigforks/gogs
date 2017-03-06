package repo

import (
	"strings"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/context"

	"github.com/gogits/gogs/modules/log"
)

// CollaborationOrgPost adds a collaboration between an Itsyou.Online organization
// and a repository
func CollaborationOrgPost(ctx *context.Context) {
	name := strings.ToLower(ctx.Query("organization"))
	if len(name) == 0 || ctx.Repo.Owner.LowerName == name {
		log.Warn("Guess what, there aint no name in the query")
		ctx.Redirect(ctx.Repo.RepoLink + "/settings/collaboration")
		return
	}

	if err := ctx.Repo.Repository.AddIyoCollaborator(name); err != nil {
		ctx.Handle(500, "AddCollaborator", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("repo.settings.add_collaborator_success"))
	ctx.Redirect(ctx.Repo.RepoLink + "/settings/collaboration")
}

// ChangeIyoCollaborationAccessMode changes the access mode of an organization to
// an Itsyou.Online repository
func ChangeIyoCollaborationAccessMode(ctx *context.Context) {
	if err := ctx.Repo.Repository.ChangeIyoCollaborationAccessMode(
		ctx.Query("uid"),
		models.AccessMode(ctx.QueryInt("mode"))); err != nil {
		log.Error(4, "ChangeIyoCollaborationAccessMode: %v", err)
	}
}

// DeleteIyoCollaboration removes an Itsyou.Online organization collaboration from
// a repository
func DeleteIyoCollaboration(ctx *context.Context) {
	if err := ctx.Repo.Repository.DeleteIyoCollaboration(ctx.Query("id")); err != nil {
		ctx.Flash.Error("DeleteIyoCollaboration: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.remove_collaborator_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": ctx.Repo.RepoLink + "/settings/collaboration",
	})
}
