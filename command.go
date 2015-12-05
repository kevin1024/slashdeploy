package slashdeploy

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ejholmes/slash"
	"github.com/ejholmes/slashdeploy/deployments"
	"golang.org/x/net/context"
)

// Command is a slash.Handler that handles the base /deploy command.
type Command struct {
	slash.Handler
	deployments.Deployer
}

func newCommand(verificationToken string) *Command {
	c := &Command{}

	d := slash.NewMux()
	d.Match(slash.MatchSubcommand("help"), slash.HandlerFunc(c.Help))
	d.MatchText(regexp.MustCompile(`(?P<repo>\S+?)@(?P<ref>\S+?) to (?P<environment>\S+?)$`), slash.HandlerFunc(c.Deploy))
	d.MatchText(regexp.MustCompile(`(?P<repo>\S+?) to (?P<environment>\S+?)$`), slash.HandlerFunc(c.Deploy))
	d.MatchText(regexp.MustCompile(`(?P<repo>\S+?)@(?P<ref>\S+?)$`), slash.HandlerFunc(c.Deploy))
	d.MatchText(regexp.MustCompile(`(?P<repo>\S+?)$`), slash.HandlerFunc(c.Deploy))

	r := slash.NewMux()
	r.Command("/deploy", verificationToken, d)

	c.Handler = r

	return c
}

// Help handles the /deploy help subcommand.
func (c *Command) Help(ctx context.Context, r slash.Responder, command slash.Command) (slash.Response, error) {
	return slash.Reply(helpText), nil
}

func (c *Command) Deploy(ctx context.Context, r slash.Responder, command slash.Command) (slash.Response, error) {
	params := slash.Params(ctx)
	owner, repo, err := splitRepo(params["repo"])
	if err != nil {
		return slash.NoResponse, err
	}

	environment := params["environment"]
	if environment == "" {
		environment = DefaultEnvironment
	}

	ref := params["ref"]
	if ref == "" {
		ref = DefaultRef
	}

	_, err = c.Deployer.Deploy(deployments.DeploymentRequest{
		Owner:       owner,
		Repository:  repo,
		Environment: environment,
		Ref:         ref,
	}, nil)

	return slash.Say(fmt.Sprintf("Created deployment request for %s/%s@%s to %s", owner, repo, ref, environment)), nil
}

var errInvalidRepo = errors.New("repo not valid")

func splitRepo(fullName string) (repo, owner string, err error) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		err = errInvalidRepo
		return
	}

	repo, owner = parts[0], parts[1]
	return
}

var helpText = `To deploy a repo to the default environment: /deploy REPO
To deploy a repo to a specific environment: /deploy REPO to ENVIRONMENT
To deploy a repo to a specific branch: /deploy REPO@REF to ENVIRONMENT`