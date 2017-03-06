package models

import (
	"fmt"

	"github.com/gogits/gogs/modules/log"
)

type IyoCollaboration struct {
	ID                   int64      `xorm:"pk autoincr"`
	RepoID               int64      `xorm:"UNIQUE(s) INDEX NOT NULL"`
	OrganizationGlobalId string     `xorm:"UNIQUE(s) INDEX NOT NULL"`
	Mode                 AccessMode `xorm:"DEFAULT 2 NOT NULL"`
}

// IyoCollaborator represents an itsyou.online organization with collaboration details.
type IyoCollaborator struct {
	OrganizationGlobalId string
	IyoCollaboration     *IyoCollaboration
}

// ModeI18nKey loads and returns a translation for the UI
func (c *IyoCollaboration) ModeI18nKey() string {
	switch c.Mode {
	case ACCESS_MODE_READ:
		return "repo.settings.collaboration.read"
	case ACCESS_MODE_WRITE:
		return "repo.settings.collaboration.write"
	case ACCESS_MODE_ADMIN:
		return "repo.settings.collaboration.admin"
	default:
		return "repo.settings.collaboration.undefined"
	}
}

// AddIyoCollaborator adds new Iyo organization collaboration to a repository with default access mode.
func (repo *Repository) AddIyoCollaborator(globalID string) error {
	collaboration := &IyoCollaboration{
		RepoID:               repo.ID,
		OrganizationGlobalId: globalID,
	}

	has, err := x.Get(collaboration)
	if err != nil {
		return err
	} else if has {
		return nil
	}
	collaboration.Mode = ACCESS_MODE_WRITE

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.InsertOne(collaboration); err != nil {
		return err
	}

	if repo.Owner.IsOrganization() {
		err = repo.recalculateTeamAccesses(sess, 0)
	} else {
		err = repo.recalculateAccesses(sess)
	}
	if err != nil {
		return fmt.Errorf("recalculateAccesses 'team=%v': %v", repo.Owner.IsOrganization(), err)
	}

	return sess.Commit()
}

// GetIyoCollaborators returns the itsyou.online organization collaborators for a repository
func (repo *Repository) GetIyoCollaborators() ([]*IyoCollaborator, error) {
	return repo.getIyoCollaborators(x)
}

// ChangeIyoCollaborationAccessMode sets new access mode for the itsyou.online organization collaboration.
func (repo *Repository) ChangeIyoCollaborationAccessMode(globalID string, mode AccessMode) error {
	// Discard invalid input
	if mode <= ACCESS_MODE_NONE || mode > ACCESS_MODE_OWNER {
		return nil
	}

	collaboration := &IyoCollaboration{
		RepoID:               repo.ID,
		OrganizationGlobalId: globalID,
	}
	has, err := x.Get(collaboration)
	if err != nil {
		return fmt.Errorf("get collaboration: %v", err)
	} else if !has {
		return nil
	}

	if collaboration.Mode == mode {
		return nil
	}
	collaboration.Mode = mode

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Id(collaboration.ID).AllCols().Update(collaboration); err != nil {
		return fmt.Errorf("update collaboration: %v", err)
	} else if _, err = sess.Exec("UPDATE iyo_access SET mode = ? WHERE organization_global_id = ? AND repo_id = ?", mode, globalID, repo.ID); err != nil {
		return fmt.Errorf("update access table: %v", err)
	}

	return sess.Commit()
}

// DeleteIyoCollaboration removes the itsyou.online organization collaboration relation
// between the user and repository.
func (repo *Repository) DeleteIyoCollaboration(globalID string) (err error) {
	log.Warn("ORG NAME: ", globalID)
	collaboration := &IyoCollaboration{
		RepoID:               repo.ID,
		OrganizationGlobalId: globalID,
	}

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if has, err := sess.Delete(collaboration); err != nil || has == 0 {
		return err
	} else if err = repo.recalculateAccesses(sess); err != nil {
		return err
	}

	return sess.Commit()
}

func (repo *Repository) getIyoCollaborations(e Engine) ([]*IyoCollaboration, error) {
	collaborations := make([]*IyoCollaboration, 0)
	return collaborations, e.Find(&collaborations, &IyoCollaboration{RepoID: repo.ID})
}

func (repo *Repository) getIyoCollaborators(e Engine) ([]*IyoCollaborator, error) {
	collaborations, err := repo.getIyoCollaborations(e)
	if err != nil {
		return nil, fmt.Errorf("getIyoCollaborations: %v", err)
	}

	collaborators := make([]*IyoCollaborator, len(collaborations))
	for i, c := range collaborations {
		globalID := c.OrganizationGlobalId
		collaborators[i] = &IyoCollaborator{
			OrganizationGlobalId: globalID,
			IyoCollaboration:     c,
		}
	}
	return collaborators, nil
}
