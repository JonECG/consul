package policycreate

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/consul/agent"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/logger"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/hashicorp/consul/testrpc"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/require"
)

func TestPolicyCreateCommand_noTabs(t *testing.T) {
	t.Parallel()

	if strings.ContainsRune(New(cli.NewMockUi()).Help(), '\t') {
		t.Fatal("help has tabs")
	}
}

func TestPolicyCreateCommand(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	testDir := testutil.TempDir(t, "acl")
	defer os.RemoveAll(testDir)

	a := agent.NewTestAgent(t, t.Name(), `
	primary_datacenter = "dc1"
	acl {
		enabled = true
		tokens {
			master = "root"
		}
	}`)

	a.Agent.LogWriter = logger.NewLogWriter(512)

	defer a.Shutdown()
	testrpc.WaitForLeader(t, a.RPC, "dc1")

	ui := cli.NewMockUi()
	cmd := New(ui)

	rules := []byte("service \"\" { policy = \"write\" }")
	err := ioutil.WriteFile(testDir+"/rules.hcl", rules, 0644)
	require.NoError(err)

	t.Run("Basic Policy", func(t *testing.T) {
		args := []string{
			"-http-addr=" + a.HTTPAddr(),
			"-token=root",
			"-name=foobar",
			"-rules=@" + testDir + "/rules.hcl",
		}

		code := cmd.Run(args)
		require.Equal(code, 0)
		require.Empty(ui.ErrorWriter.String())
	})

	t.Run("Policy with ID", func(t *testing.T) {
		args := []string{
			"-http-addr=" + a.HTTPAddr(),
			"-token=root",
			"-name=id-init",
			"-id=6ac6a30a-d84e-4257-a149-ab652f3f04b9",
			"-rules=@" + testDir + "/rules.hcl",
		}

		code := cmd.Run(args)
		require.Empty(ui.ErrorWriter.String())
		require.Equal(code, 0)

		conf := api.DefaultConfig()
		conf.Address = a.HTTPAddr()
		conf.Token = "root"

		// going to use the API client to grab the token - we could potentially try to grab the values
		// out of the command output but this seems easier.
		client, err := api.NewClient(conf)
		require.NoError(err)
		require.NotNil(client)

		policy, _, err := client.ACL().PolicyRead("6ac6a30a-d84e-4257-a149-ab652f3f04b9", nil)
		require.NoError(err)
		require.Equal("6ac6a30a-d84e-4257-a149-ab652f3f04b9", policy.ID)
	})
}
