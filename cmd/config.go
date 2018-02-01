package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func stringConfig(cmd *cobra.Command, name, short, value, description string) {
	cmd.PersistentFlags().StringP(name, short, value, description)
	viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
	viper.BindEnv(name, strings.ToUpper(name))
}
