package gitee

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/antihax/optional"

	"github.com/koderover/zadig/pkg/tool/git"
	"github.com/koderover/zadig/pkg/tool/httpclient"
)

const (
	GiteeHOSTURL = "https://gitee.com/api"
)

type Project struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch,omitempty"`
}

func (c *Client) ListRepositoriesForAuthenticatedUser(accessToken, keyword string, page, perPage int) ([]Project, error) {
	httpClient := httpclient.New(
		httpclient.SetHostURL(GiteeHOSTURL),
	)
	url := "/v5/user/repos"
	queryParams := make(map[string]string)
	queryParams["access_token"] = accessToken
	queryParams["visibility"] = "all"
	queryParams["affiliation"] = "owner"
	queryParams["q"] = keyword
	queryParams["page"] = strconv.Itoa(page)
	queryParams["per_page"] = strconv.Itoa(perPage)

	var projects []Project
	_, err := httpClient.Get(url, httpclient.SetQueryParams(queryParams), httpclient.SetResult(&projects))
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *Client) ListRepositoriesForOrg(accessToken, org string, page, perPage int) ([]Project, error) {
	httpClient := httpclient.New(
		httpclient.SetHostURL(GiteeHOSTURL),
	)
	url := fmt.Sprintf("/v5/orgs/%s/repos", org)
	queryParams := make(map[string]string)
	queryParams["access_token"] = accessToken
	queryParams["type"] = "all"
	queryParams["page"] = strconv.Itoa(page)
	queryParams["per_page"] = strconv.Itoa(perPage)

	var projects []Project
	_, err := httpClient.Get(url, httpclient.SetQueryParams(queryParams), httpclient.SetResult(&projects))
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *Client) ListHooks(ctx context.Context, owner, repo string, opts *gitee.GetV5ReposOwnerRepoHooksOpts) ([]gitee.Hook, error) {
	hs, _, err := c.WebhooksApi.GetV5ReposOwnerRepoHooks(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	return hs, nil
}

func (c *Client) DeleteHook(ctx context.Context, owner, repo string, id int64) error {
	_, err := c.WebhooksApi.DeleteV5ReposOwnerRepoHooksId(ctx, owner, repo, int32(id), nil)
	if err != nil {
		return err
	}
	return nil
}

type Hook struct {
	ID                  int       `json:"id"`
	URL                 string    `json:"url"`
	CreatedAt           time.Time `json:"created_at"`
	Password            string    `json:"password"`
	ProjectID           int       `json:"project_id"`
	Result              string    `json:"result"`
	ResultCode          int       `json:"result_code"`
	PushEvents          bool      `json:"push_events"`
	TagPushEvents       bool      `json:"tag_push_events"`
	IssuesEvents        bool      `json:"issues_events"`
	NoteEvents          bool      `json:"note_events"`
	MergeRequestsEvents bool      `json:"merge_requests_events"`
}

func (c *Client) CreateHook(accessToken, owner, repo string, hook *git.Hook) (*Hook, error) {
	httpClient := httpclient.New(
		httpclient.SetHostURL(GiteeHOSTURL),
	)
	url := fmt.Sprintf("/v5/repos/%s/%s/hooks", owner, repo)
	var hookInfo *Hook
	_, err := httpClient.Post(url, httpclient.SetBody(struct {
		AccessToken         string `json:"access_token"`
		URL                 string `json:"url"`
		Password            string `json:"password"`
		PushEvents          string `json:"push_events"`
		TagPushEvents       string `json:"tag_push_events"`
		MergeRequestsEvents string `json:"merge_requests_events"`
	}{accessToken, hook.URL, hook.Secret, "true", "true", "true"}), httpclient.SetResult(&hookInfo))
	if err != nil {
		return nil, err
	}
	return hookInfo, nil
}

func (c *Client) UpdateHook(ctx context.Context, owner, repo string, id int64, hook *git.Hook) (gitee.Hook, error) {
	resp, _, err := c.WebhooksApi.PatchV5ReposOwnerRepoHooksId(ctx, owner, repo, int32(id), hook.URL, &gitee.PatchV5ReposOwnerRepoHooksIdOpts{
		Password:            optional.NewString(hook.Secret),
		PushEvents:          optional.NewBool(true),
		TagPushEvents:       optional.NewBool(true),
		MergeRequestsEvents: optional.NewBool(true),
	})
	if err != nil {
		return gitee.Hook{}, err
	}

	return resp, nil
}

func (c *Client) GetContents(ctx context.Context, owner, repo, sha string) (gitee.Blob, error) {
	fileContent, _, err := c.GitDataApi.GetV5ReposOwnerRepoGitBlobsSha(ctx, owner, repo, sha, &gitee.GetV5ReposOwnerRepoGitBlobsShaOpts{})
	if err != nil {
		return gitee.Blob{}, err
	}
	return fileContent, nil
}

// "Recursive" Assign a value of 1 to get the directory recursively
// sha Can be a branch name (such as master), Commit, or the SHA value of the directory Tree
func (c *Client) GetTrees(ctx context.Context, owner, repo, sha string, level int) (gitee.Tree, error) {
	tree, _, err := c.GitDataApi.GetV5ReposOwnerRepoGitTreesSha(ctx, owner, repo, sha, &gitee.GetV5ReposOwnerRepoGitTreesShaOpts{
		Recursive: optional.NewInt32(int32(level)),
	})
	if err != nil {
		return gitee.Tree{}, err
	}
	return tree, nil
}

type RepoCommit struct {
	URL    string `json:"url"`
	Sha    string `json:"sha"`
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Date  time.Time `json:"date"`
			Email string    `json:"email"`
		} `json:"author"`
		Committer struct {
			Name  string    `json:"name"`
			Date  time.Time `json:"date"`
			Email string    `json:"email"`
		} `json:"committer"`
		Message string `json:"message"`
	} `json:"commit"`
}

func (c *Client) GetSingleCommitOfProject(ctx context.Context, accessToken, owner, repo, commitSha string) (*RepoCommit, error) {
	httpClient := httpclient.New(
		httpclient.SetHostURL(GiteeHOSTURL),
	)
	url := fmt.Sprintf("/v5/repos/%s/%s/commits/%s", owner, repo, commitSha)

	var commit *RepoCommit
	_, err := httpClient.Get(url, httpclient.SetQueryParam("access_token", accessToken), httpclient.SetResult(&commit))
	if err != nil {
		return nil, err
	}
	return commit, nil
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int    `json:"created_at"`
}

func RefreshAccessToken(refreshToken string) (*AccessToken, error) {
	httpClient := httpclient.New(
		httpclient.SetHostURL("https://gitee.com"),
	)
	url := "/oauth/token"
	queryParams := make(map[string]string)
	queryParams["grant_type"] = "refresh_token"
	queryParams["refresh_token"] = refreshToken

	var accessToken *AccessToken
	_, err := httpClient.Post(url, httpclient.SetQueryParams(queryParams), httpclient.SetResult(&accessToken))
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

type Compare struct {
	Commits []struct {
		URL         string `json:"url"`
		Sha         string `json:"sha"`
		HTMLURL     string `json:"html_url"`
		CommentsURL string `json:"comments_url"`
		Commit      struct {
			Author struct {
				Name  string    `json:"name"`
				Date  time.Time `json:"date"`
				Email string    `json:"email"`
			} `json:"author"`
			Committer struct {
				Name  string    `json:"name"`
				Date  time.Time `json:"date"`
				Email string    `json:"email"`
			} `json:"committer"`
			Message string `json:"message"`
			Tree    struct {
				Sha string `json:"sha"`
				URL string `json:"url"`
			} `json:"tree"`
		} `json:"commit"`
		Author struct {
			ID                int    `json:"id"`
			Login             string `json:"login"`
			Name              string `json:"name"`
			AvatarURL         string `json:"avatar_url"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			Remark            string `json:"remark"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
		} `json:"author"`
		Committer struct {
			ID                int    `json:"id"`
			Login             string `json:"login"`
			Name              string `json:"name"`
			AvatarURL         string `json:"avatar_url"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			Remark            string `json:"remark"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
		} `json:"committer"`
		Parents []struct {
			Sha string `json:"sha"`
			URL string `json:"url"`
		} `json:"parents"`
	} `json:"commits"`
	Files []struct {
		Sha        string `json:"sha"`
		Filename   string `json:"filename"`
		Status     string `json:"status"`
		Additions  int    `json:"additions"`
		Deletions  int    `json:"deletions"`
		Changes    int    `json:"changes"`
		BlobURL    string `json:"blob_url"`
		RawURL     string `json:"raw_url"`
		ContentURL string `json:"content_url"`
		Patch      string `json:"patch"`
	} `json:"files"`
}

func (c *Client) GetReposOwnerRepoCompareBaseHead(accessToken, owner string, repo string, base string, head string) (*Compare, error) {
	httpClient := httpclient.New(
		httpclient.SetHostURL(GiteeHOSTURL),
	)
	url := fmt.Sprintf("/v5/repos/%s/%s/compare/%s...%s", owner, repo, base, head)

	var compare *Compare
	_, err := httpClient.Get(url, httpclient.SetQueryParam("access_token", accessToken), httpclient.SetResult(&compare))
	if err != nil {
		return nil, err
	}
	return compare, nil
}
