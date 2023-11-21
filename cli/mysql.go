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

type dumpCmd struct {
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
		newDumpCmd(config).cmd,
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
			mysql := &infrastructure.SQL{}
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

func newDumpCmd(config *viper.Viper) *dumpCmd {
	dump := &dumpCmd{}
	cmd := &cobra.Command{
		Use:                   "dump",
		Short:                 "Dump database",
		Long:                  strings.TrimSpace("Dump Database blah blah"),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			mysql := &infrastructure.SQL{}
			context := infrastructure.NewContext(mysql)
			_, err := context.ConnectToDatabase(config)
			if err != nil {
				slog.Error("Panic error", err)
			} else {
				slog.Debug(fmt.Sprintf("Connected to Database: %s", config.GetString("mysql-database")))
			}
			mysql.Dump()
			return nil
		},
	}

	dump.cmd = cmd
	return dump
}
