package cmd

import (
	"github.com/gork74/aws-parameter-bulk/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() { // nolint: gochecknoinits
	saveCmd := &cobra.Command{
		Args:  cobra.MinimumNArgs(1),
		Use:   "save [file] [basepath]",
		Short: "save .env",
		Long: "save .env\n" +
			"save .env /basepath\n\n" +
			"saves each entry from a file in the .env format (KEY=value) into multiple variables in the form key=value\n" +
			"or saves them into multiple variables in the form /basepath/key=value",
		Run: func(cmd *cobra.Command, args []string) {
			fileName := args[0]
			path := ""
			if len(args) > 1 {
				path = args[1]
			}
			inJsonFlag, _ := cmd.Flags().GetBool("injson")
			dryFlag, _ := cmd.Flags().GetBool("dry")
			log.Debug().Msgf("Filename: %s Path: %s", fileName, path)
			log.Debug().Msgf("Input JSON: %t", inJsonFlag)
			log.Debug().Msgf("Dry: %v", dryFlag)

			ssmClient := util.NewSSM()
			err := ssmClient.SaveParametersFromFile(fileName, path, inJsonFlag, dryFlag)
			if err != nil {
				log.Error().Msg(err.Error())
				return
			}
		},
	}
	saveCmd.PersistentFlags().Bool("injson", false, "Parse input file as json and extract each json value as output.")
	saveCmd.PersistentFlags().Bool("dry", false, "Dry run, just output what would be saved to ssm and do nothing.")
	rootCmd.AddCommand(saveCmd)
}
