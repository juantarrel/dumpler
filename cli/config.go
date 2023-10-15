package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"strings"
)

type configureCmd struct {
	cmd      *cobra.Command
	host     string
	port     string
	username string
	password string
	database string
}

func newConfigureCmd(config *viper.Viper) *configureCmd {
	configureStruct := &configureCmd{}
	cmd := &cobra.Command{
		Use:                   "configure",
		Short:                 "Configure Database",
		Long:                  strings.TrimSpace("Configure Database - blah blah blah"),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config.Set("mysql-host", configureStruct.host)
			config.Set("mysql-port", configureStruct.port)
			config.Set("mysql-username", configureStruct.username)
			config.Set("mysql-password", configureStruct.password)
			config.Set("mysql-database", configureStruct.database)
			//config.WriteConfigAs("config.yaml")
			err := config.WriteConfig()
			if err != nil {
				fmt.Println(err)
				slog.Debug("Error: ", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configureStruct.host, "host", "s", "localhost", "Host")
	cmd.Flags().StringVarP(&configureStruct.port, "port", "o", "3306", "Port")
	cmd.Flags().StringVarP(&configureStruct.username, "username", "u", "", "Username")
	cmd.Flags().StringVarP(&configureStruct.password, "password", "p", "", "Password")
	cmd.Flags().StringVarP(&configureStruct.database, "database", "d", "", "Database")

	configureStruct.cmd = cmd
	return configureStruct
}
