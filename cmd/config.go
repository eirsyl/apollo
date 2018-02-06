package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"

	log "github.com/sirupsen/logrus"
)

func stringConfig(cmd *cobra.Command, name, short, value, description string) {
	cmd.PersistentFlags().StringP(name, short, value, description)
	err := viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
	if err != nil {
		log.Warnf("Could not bind flag to viper: %v", err)
	}
	err = viper.BindEnv(name, strings.ToUpper(name))
	if err != nil {
		log.Warnf("Could not bind viper value to env: %v", err)
	}
}
