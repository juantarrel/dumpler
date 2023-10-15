package cli

import (
	"fmt"
	"github.com/juantarrel/dumpler/infrastructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"strings"
)

type mysqlCmd struct {
	cmd *cobra.Command
}

type pingCmd struct {
	cmd *cobra.Command
}

func newMysqlCmd(config *viper.Viper) *mysqlCmd {
	mysqlStruct := &mysqlCmd{}
	cmd := &cobra.Command{
		Use:                   "mysql",
		Short:                 "Manage mysql command",
		Long:                  strings.TrimSpace("Manage all mysql commands - blah blah blah"),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newPingCmd(config).cmd,
	)
	mysqlStruct.cmd = cmd
	return mysqlStruct
}

func newPingCmd(config *viper.Viper) *pingCmd {
	ping := &pingCmd{}
	cmd := &cobra.Command{
		Use:                   "ping",
		Short:                 "Ping Database",
		Long:                  strings.TrimSpace("Ping Database blah blah"),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mysql := &infrastructure.MySQL{}
			context := infrastructure.NewContext(mysql)
			db, _ := context.ConnectToDatabase(config)
			err := db.Ping()
			if err != nil {
				slog.Error("Panic error", err)
			} else {
				slog.Debug("Ping was success")
				fmt.Println("Ping Success")
			}

			return nil
		},
	}
	ping.cmd = cmd
	return ping
}
