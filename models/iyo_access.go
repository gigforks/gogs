package models

import (
	"github.com/gogits/gogs/authorization"
)

// IyoAccessLevel gets the IYO organization collaborators for the repository. If the
// organization has a higher access level than the already acquired one, make an api call
// to Itsyou.Online to check if the user is in the organization, and adjust the
// access level accordingly
func IyoAccessLevel(e Engine, u *User, repo *Repository, acquiredAccess AccessMode) (AccessMode, error) {
	mode := acquiredAccess
	orgAccess := make([]*IyoCollaboration, 0)
	err := e.Where("repo_id=?", repo.ID).Find(&orgAccess)
	if err != nil {
		return acquiredAccess, err
	}

	for _, a := range orgAccess {
		// If this organization doesn't have an access mode that is higher then the one
		// we already confirmed, skip to the next organization.
		if a.Mode <= mode {
			continue
		}
		// If this organization has a higher access mode than the one we already acquired,
		// check if the user is part of the organization
		if authorization.UserIsMemberOfOrg(u.Name, a.OrganizationGlobalId) {
			mode = a.Mode
		}

	}

	return mode, nil
}
