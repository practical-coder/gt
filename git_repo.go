package gt

import (
	"fmt"

	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/rs/zerolog"
	ssh "golang.org/x/crypto/ssh"
)

var Logger zerolog.Logger
var once sync.Once

func init() {
	once.Do(func() {
		Logger = zerolog.New(zerolog.NewConsoleWriter()).With().Logger()
	})
}

type GitRepo struct {
	URL        string
	Path       string
	BranchName string
	Repo       *git.Repository
	PublicKeys *gitssh.PublicKeys
}

func (gr *GitRepo) SetKeys(privateKey []byte) error {
	key, err := ssh.ParseRawPrivateKey(privateKey)

	if err != nil {
		return fmt.Errorf("ParseRawPrivateKey Error: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return fmt.Errorf("NewSignerFromKey Error: %v", err)
	}

	publicKeys := gitssh.PublicKeys{
		User:   "git",
		Signer: signer,
	}
	gr.PublicKeys = &publicKeys

	return nil
}

func New(repoURL, branchName, repoPath string) (*GitRepo, error) {
	gitRepo := &GitRepo{
		URL:        repoURL,
		Path:       repoPath,
		BranchName: branchName,
	}

	return gitRepo, nil
}

func (gr *GitRepo) Open() {
	repository, err := git.PlainOpen(gr.Path)
	if err != nil {
		Logger.Info().Err(err).Msg("Repo Open Error")
		return
	}
	rems, err := repository.Remotes()
	if err == nil && len(rems) > 0 {
		rem := rems[0]
		Logger.Info().Strs("URLs", rem.Config().URLs).Msgf("Local Remote URLs Format")
	}
	gr.Repo = repository
}

func (gr *GitRepo) RevisionExists(revName string) bool {
	_, err := gr.Repo.ResolveRevision(plumbing.Revision(revName))
	if err == nil {
		return true
	}
	return false
}

func (gr *GitRepo) LatestSHA(length int) string {
	h, err := gr.Repo.ResolveRevision(plumbing.Revision(gr.BranchName))
	if err != nil {
		Logger.Info().Err(err).Msg("LatestSHA Error")
	}
	return h.String()[:length]
}

func (gr GitRepo) Worktree() *git.Worktree {
	workTree, err := gr.Repo.Worktree()
	if err != nil {
		Logger.Info().Err(err).Msg("Repo Worktree Error")
		return nil
	}
	return workTree
}

func (gr *GitRepo) Pull() error {
	gr.Open()
	if gr.Repo == nil {
		return fmt.Errorf("Git Repository: Open Error. Path: %s", gr.Path)
	}

	w := gr.Worktree()
	if w == nil {
		return fmt.Errorf("Git Repository: Worktree Error. Path: %s", gr.Path)
	}

	err := w.Pull(gr.PullOptions())

	if err != nil {
		return err
	}

	return nil
}

func (gr *GitRepo) ReferenceName() plumbing.ReferenceName {
	refName := fmt.Sprintf("refs/heads/%s", gr.BranchName)
	return plumbing.ReferenceName(refName)
}

func (gr *GitRepo) CloneOptions() *git.CloneOptions {
	cloneOptions := &git.CloneOptions{
		Auth:          gr.PublicKeys,
		URL:           gr.URL,
		ReferenceName: gr.ReferenceName(),
		Depth:         10,
		Progress:      os.Stdout,
	}

	return cloneOptions
}

func (gr *GitRepo) PullOptions() *git.PullOptions {
	pullOptions := &git.PullOptions{
		Auth:          gr.PublicKeys,
		RemoteName:    "origin",
		ReferenceName: gr.ReferenceName(),
	}

	return pullOptions
}

func (gr *GitRepo) EnsurePath() error {
	_, err := os.Stat(gr.Path)
	if os.IsNotExist(err) {
		err = os.Mkdir(gr.Path, 0755)
		if err != nil {
			Logger.Info().
				Err(err).
				Str("repo_path", gr.Path).
				Msg("Mkdir Error on repository dir path")
		}
	} else {
		return err
	}

	return nil
}

func (gr *GitRepo) Clone() error {
	err := gr.EnsurePath()
	if err != nil {
		return err
	}

	repository, err := git.PlainClone(gr.Path, false, gr.CloneOptions())

	if err != nil {
		return err
	}

	gr.Repo = repository

	return nil
}

func (gr *GitRepo) CloneOrPull() error {
	err := gr.Clone()

	switch err {
	case nil:
		return nil
	case git.ErrRepositoryAlreadyExists:
		pullErr := gr.Pull()
		if pullErr != nil {
			return pullErr
		}
	default:
		return err
	}

	return nil
}
