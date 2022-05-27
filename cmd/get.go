package cmd

import (
	"fmt"
	"github.com/gork74/aws-parameter-bulk/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

func init() { // nolint: gochecknoinits
	getCmd := &cobra.Command{
		Args:  cobra.MinimumNArgs(1),
		Use:   "get [names]",
		Short: "get name1,/path1,/path2/subpath",
		Long: "get name1,/path1,/path2/subpath,name3,name4\n\n" +
			"Accepts paths and non-path names, as a colon-separated list.\n" +
			"For names, returns the name and value as name=value.\n" +
			"For paths, returns all parameter store values under a path as name=value.\n" +
			"This can be piped into an file (> .env), to be included via --env-file=.env\n" +
			"or to be set in a shell environment (not recommended): export $(cat .env).\n" +
			"Note: name output is unique, if two paths parameters have the same name, the value of the last name in the list wins\n" +
			"Use --help for help on the flags: --export --injson --outjson --upper --quote --norecursive --prefixpath --prefixnormalizedpath",
		Run: func(cmd *cobra.Command, args []string) {
			exportFlag, _ := cmd.Flags().GetBool("export")
			inJsonFlag, _ := cmd.Flags().GetBool("injson")
			outJsonFlag, _ := cmd.Flags().GetBool("outjson")
			upperFlag, _ := cmd.Flags().GetBool("upper")
			quoteFlag, _ := cmd.Flags().GetBool("quote")
			// use recursive as default, to stay backward compatible
			noRecursiveFlag, _ := cmd.Flags().GetBool("norecursive")
			recursiveFlag := !noRecursiveFlag
			prefixPathFlag, _ := cmd.Flags().GetBool("prefixpath")
			prefixNormalizedPathFlag, _ := cmd.Flags().GetBool("prefixnormalizedpath")
			flags := util.Flags{
				exportFlag,
				inJsonFlag,
				outJsonFlag,
				upperFlag,
				quoteFlag,
				false,
				recursiveFlag,
				prefixPathFlag,
				prefixNormalizedPathFlag,
			}
			log.Debug().Msgf("Names/Paths: %s", args[0])
			log.Debug().Msgf("Flags: %+v", flags)
			ssmClient := util.NewSSM()
			result, err := ssmClient.GetParams(&args[0], flags)
			if err != nil {
				log.Error().Msg(err.Error())
				os.Exit(1)
				return
			}

			output, err := ssmClient.GetOutputString(result, flags)
			if err != nil {
				log.Error().Msg(err.Error())
				os.Exit(1)
				return
			}
			fmt.Print(output)
		},
	}
	getCmd.PersistentFlags().Bool("export", false, "Prefix output with export to eval it in shell")
	getCmd.PersistentFlags().Bool("injson", false, "Parse input parameter values as json and extract each json value as output. Each has to be json.")
	getCmd.PersistentFlags().Bool("outjson", false, "Output everything as a json file. Does not make sense together with --export.")
	getCmd.PersistentFlags().Bool("upper", false, "Make keys uppercase")
	getCmd.PersistentFlags().Bool("quote", false, "Wrap values in quotes")
	getCmd.PersistentFlags().Bool("norecursive", false, "Do not read recursively if getting a path")
	getCmd.PersistentFlags().Bool("prefixpath", false, "Prefix names with the path")
	getCmd.PersistentFlags().Bool("prefixnormalizedpath", false, "Prefix names with the normalized path")
	rootCmd.AddCommand(getCmd)

}
