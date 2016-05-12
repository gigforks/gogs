package issues

import (
	"github.com/gigforks/gogs/models"
	"github.com/gigforks/gogs/modules/log"
	"github.com/gigforks/gogs/modules/middleware"
	// "github.com/gigforks/gogs/modules/auth"
	// "github.com/gigforks/gogs/modules/user"

	"github.com/gigforks/gogs/routers/api/v1/convert"
	// "github.com/gigforks/gogs/routers/repo"
	// "github.com/gigforks/gogs/routers/api/v1/user"
)

//CreateIssueOption option used to recieve post body from client 
type CreateIssueOption struct {
	RepoName    string `json:"repo"`
	Title       string `json:"title"`
	LabelIDs    string `json:"label"`
	MilestoneID string `json:"milestone"`
	AssigneeID  string `json:"assignee"`
	Content     string `json:"content"`
}

//CreateIssue function call for post request to create issue /repo/issues
func CreateIssue(ctx *middleware.Context, opt CreateIssueOption) {
	log.Debug("Inside create issue")
	u, err := models.GetUserByName(ctx.User.Name)

	if err != nil {
		ctx.Handle(500, "NewIssue", err)
		return
	}

	repo, err := models.GetRepositoryByName(u.Id, opt.RepoName)
	if err != nil {
		ctx.Handle(500, "NewIssue", err)
		return
	}

	issue := &models.Issue{
		RepoID:   repo.ID,
		Index:    repo.NextIssueIndex(),
		Name:     opt.Title,
		PosterID: u.Id,
		Poster:   u,
		// MilestoneID: opt.MilestoneID,
		// AssigneeID:  opt.AssigneeID,
		Content: opt.Content,
	}

	if err := models.NewIssue(repo, issue, nil, nil); err != nil {
		ctx.Handle(500, "NewIssue", err)
		return
	}

	log.Trace("Issue created: %d/%d", repo.ID, issue.ID)

	if ctx.Written() {
		return
	}

	log.Trace("Issue created: %d/%d", repo.ID, issue.ID)
	ctx.JSON(201, convert.ToApiUser(u))
}

func DeleteIssue() {

}

func ListIssues(ctx *middleware.Context) {
	u, err := models.GetUserByName(ctx.User.Name)
	
    if err != nil {
		ctx.Handle(500, "listuser", err)
		return
	}

    repo, err := models.GetRepositoryByName(u.Id, ctx.Params(":reponame"))

    if err != nil {
		ctx.Handle(500, "listuser", err)
		return
	}
    
    issues, err := models.Issues(&models.IssuesOptions{
		UserID:     u.Id,
		RepoID:     repo.ID,
	})
    
    if err != nil {
        ctx.Handle(500, "listuser", err)
		return
    }    
    
    
	if ctx.Written() {
		return
	}


	ctx.JSON(200, convert.ToApiUser(u))

}

func Get() {

}

func Edit() {

}
