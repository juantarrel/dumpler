package cli

import (
	"errors"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/juantarrel/dumpler/fs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
)

type rootCmd struct {
	cmd     *cobra.Command
	exit    func(int)
	verbose bool
}

func createViperConfig() (*viper.Viper, error) {
	config := viper.New()
	appConfigPath, err := fs.GetAppConfigPath()
	if err != nil {
		return nil, err
	}
	config.AddConfigPath(appConfigPath)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	return config, nil
}

func Execute(args []string) {
	viperConfig, err := createViperConfig()
	if err != nil {
		slog.Error("Failed to create viper config", "error", err)
		os.Exit(1)
	}
	newRootCmd(os.Exit, viperConfig).Execute(args)
}

func (r *rootCmd) Execute(args []string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Panic occurred", "error", err)
		}
	}()

	// Set args for root command
	r.cmd.SetArgs(args)

	if err := r.cmd.Execute(); err != nil {
		// Defaults
		code := 1
		msg := "command failed"

		// Override defaults if possible
		exitErr := &exitError{}
		if errors.As(err, &exitErr) {
			code = exitErr.code
			if exitErr.details != "" {
				msg = exitErr.details
			}
		}

		// Log error with details and exit
		slog.Debug(msg, "error", err)
		r.exit(code)
		return
	}
	r.exit(0)
}

func newRootCmd(exit func(int), config *viper.Viper) *rootCmd {
	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:   "dumpler [mysql] [file-or-db]",
		Short: "Tool to dump sql databases",
		Long:  "Tool to dump sql databases",
		Example: `
Example blah blah blah
`,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.RangeArgs(0, 2),
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if root.verbose {
				opts := &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}
				handler := slog.NewTextHandler(os.Stdout, opts)
				slog.SetDefault(slog.New(handler))
			}
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return loadViperConfig(config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			// clients (mysql, postgresql, sql server)

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false,
		"enable more verbose output for debugging")

	cc.Init(&cc.Config{
		RootCmd:       cmd,
		Headings:      cc.HiCyan + cc.Bold + cc.Underline,
		CmdShortDescr: cc.HiMagenta,
		Commands:      cc.HiYellow + cc.Bold,
		Example:       cc.Italic + cc.Bold + cc.White,
		ExecName:      cc.Bold,
		Flags:         cc.HiYellow + cc.Bold,
		FlagsDataType: cc.HiYellow + cc.Bold + cc.Green,
		FlagsDescr:    cc.HiCyan,
	})
	root.cmd = cmd

	return root
}

func loadViperConfig(config *viper.Viper) error {
	if !viper.IsSet("TESTING") {
		slog.Debug("Loading config")
		err := setViperDefaults(config)
		if err != nil {
			return err
		}
	}
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			slog.Debug("Config file not found - using defaults")
			return nil
		}
		// Config file was found but another error was produced
		return err
	}
	slog.Debug("Config file loaded")
	return nil
}

func setViperDefaults(config *viper.Viper) error {
	// cache dir
	appCacheDir, err := fs.GetAppCacheDir()
	if err != nil {
		return err
	}
	config.SetDefault("cacheDir", appCacheDir)

	return nil
}
