package v1

import (
	"github.com/go-macaron/binding"
	api "github.com/gogits/go-gogs-client"
	"github.com/gogits/gogs/routers/api/v1/org"
	"github.com/gogits/gogs/routers/api/v1/repo"
	"github.com/gogits/gogs/routers/api/v1/user"
	macaron "gopkg.in/macaron.v1"
)

func registerExtendedRoutes(m *macaron.Macaron) {
	bind := binding.Bind

	m.Group("/users", func() {
		m.Get("", user.ListAllUsers)
	})

	m.Post("/user/org", reqToken(), bind(api.CreateOrgOption{}), user.AddMyUserToOrganization)
	m.Post("/users/:username/org", reqToken(), bind(api.CreateOrgOption{}), user.AddUserToOrganization)
	m.Delete("/user/org", reqToken(), bind(api.CreateOrgOption{}), user.DeleteMyUserFromOrganization)
	m.Delete("/users/:username/org", reqToken(), bind(api.CreateOrgOption{}), user.DeleteUserFromOrganization)

	m.Delete("/collaborators/:collaborator", bind(api.AddCollaboratorOption{}), repo.DeleteCollaborator)

	m.Get("/orgs/", reqToken(), org.ListAllOrgs)
	m.Post("/orgs", reqToken(), bind(api.CreateOrgOption{}), org.CreateOrganization)

	m.Get("/orgs/search", repo.SearchOrgs)

}
