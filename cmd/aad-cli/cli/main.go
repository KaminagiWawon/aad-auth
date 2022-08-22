package cli

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubuntu/aad-auth/internal/cache"
	"github.com/ubuntu/aad-auth/internal/consts"
	"github.com/ubuntu/aad-auth/internal/logger"
)

// App encapsulates commands and options of the application.
type App struct {
	rootCmd cobra.Command
	ctx     context.Context
	domain  string
	cache   *cache.Cache

	options options
}

// options are the configurable functional options of the application.
type options struct {
	editor       string
	configFile   string
	dpkgQueryCmd string
	cache        *cache.Cache
}
type option func(*options)

// WithDpkgQueryCmd specifies a custom dpkg-query command to use for the user command.
// This is only used in tests.
func WithDpkgQueryCmd(p string) func(o *options) {
	return func(o *options) {
		o.dpkgQueryCmd = p
	}
}

// WithCache specifies a personalized cache object to use for the app.
// Useful in tests for overriding the default cache.
func WithCache(c *cache.Cache) func(o *options) {
	return func(o *options) {
		o.cache = c
	}
}

// WithEditor specifies a custom editor to use when editing the config file.
// Will probably only be used in tests.
func WithEditor(p string) func(o *options) {
	return func(o *options) {
		o.editor = p
	}
}

// WithConfigFile specifies a custom config file to use for the config command.
func WithConfigFile(p string) func(o *options) {
	return func(o *options) {
		o.configFile = p
	}
}

// New registers commands and returns a new App.
func New(opts ...option) *App {
	// Apply given options.
	args := options{
		editor:       getDefaultEditor(),
		configFile:   consts.DefaultConfigPath,
		dpkgQueryCmd: "dpkg-query",
	}

	for _, o := range opts {
		o(&args)
	}

	a := App{ctx: context.Background(), options: args}
	a.rootCmd = cobra.Command{
		Use:   "aad-cli [COMMAND]",
		Short: "Azure AD CLI",
		Long:  "Manage Azure AD accounts configuration",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			// Set logger parameters and attach to the context
			verbosity, _ := cmd.Flags().GetCount("verbose")
			logger.SetVerboseMode(verbosity)
			logrus.SetFormatter(&logger.LogrusFormatter{})
			a.ctx = logger.CtxWithLogger(a.ctx, logger.LogrusLogger{FieldLogger: logrus.StandardLogger()})

			return nil
		},
		SilenceErrors: true,
	}

	a.rootCmd.PersistentFlags().CountP("verbose", "v", "issue INFO (-v), DEBUG (-vv) or DEBUG with caller (-vvv) output")

	a.installUser()
	a.installConfig()
	a.installVersion()

	return &a
}

// Run executes the app.
func (a *App) Run() error {
	return a.rootCmd.Execute()
}

// Quit exits the app.
func (a *App) Quit() error {
	return nil
}

// UsageError returns if the error is a command parsing or runtime one.
func (a App) UsageError() bool {
	return !a.rootCmd.SilenceUsage
}

// SetArgs changes the root command args. Shouldn't be in general necessary apart for integration tests.
func (a *App) SetArgs(args []string) {
	a.rootCmd.SetArgs(args)
}

// RootCmd returns a copy of the root command for the app. Shouldn't be in
// general necessary apart from running generators.
func (a App) RootCmd() cobra.Command {
	return a.rootCmd
}
