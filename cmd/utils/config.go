package utils

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// StringConfig adds a string flag to a cli
func StringConfig(cmd *cobra.Command, name, short, value, description string) {
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

// BoolConfig adds a bool flag to a cli
func BoolConfig(cmd *cobra.Command, name, short string, value bool, description string) {
	cmd.PersistentFlags().BoolP(name, short, value, description)
	err := viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name))
	if err != nil {
		log.Warnf("Could not bind flag to viper: %v", err)
	}
	err = viper.BindEnv(name, strings.ToUpper(name))
	if err != nil {
		log.Warnf("Could not bind viper value to env: %v", err)
	}
}
