package datastore

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromConfigFile(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetHomeDir("/root/.porter")

	c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

	c.DataLoader = FromConfigFile
	err := c.LoadData()
	require.NoError(t, err, "dataloader failed")
	require.NotNil(t, c.Data, "config.Data was not populated")
	assert.True(t, c.Debug, "config.Debug was not set correctly")
}

func TestFromFlagsThenEnvVarsThenConfigFile(t *testing.T) {
	buildCommand := func(c *config.Config) *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().BoolVar(&c.Debug, "debug", false, "debug")
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return c.LoadData()
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
		c.DataLoader = FromFlagsThenEnvVarsThenConfigFile(cmd)
		return cmd
	}

	t.Run("no flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()
		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.False(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug flag", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.True(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug flag overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug=false"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.False(t, c.Debug, "config.Debug should have been set by the flag and not the config")
	})

	t.Run("debug env var", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		os.Setenv("PORTER_DEBUG", "true")
		defer os.Unsetenv("PORTER_DEBUG")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.True(t, c.Debug, "config.Debug was not set correctly")
	})

	t.Run("debug env var overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		os.Setenv("PORTER_DEBUG", "false")
		defer os.Unsetenv("PORTER_DEBUG")
		c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.False(t, c.Debug, "config.Debug should have been set by the env var and not the config")
	})

	t.Run("flag overrides debug env var overrides config", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.SetHomeDir("/root/.porter")
		os.Setenv("PORTER_DEBUG", "false")
		defer os.Unsetenv("PORTER_DEBUG")
		c.TestContext.AddTestFile("testdata/config.toml", "/root/.porter/config.toml")

		cmd := buildCommand(c.Config)
		cmd.SetArgs([]string{"--debug", "true"})
		err := cmd.Execute()

		require.NoError(t, err, "dataloader failed")
		require.NotNil(t, c.Data, "config.Data was not populated")
		assert.True(t, c.Debug, "config.Debug should have been set by the flag and not the env var or config")
	})
}
