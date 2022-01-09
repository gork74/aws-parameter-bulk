package cmd

import (
	"github.com/gork74/aws-parameter-bulk/server"

	"github.com/spf13/cobra"
)

// Web ui command
func init() { // nolint: gochecknoinits
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "web ",
		Long: "web \n\n" +
			"Starts a web ui on localhost:8888 which can show, compare and edit multiple ssm entries",
		Run: func(cmd *cobra.Command, args []string) {
			address, _ := cmd.Flags().GetString("address")
			if address == "" {
				address = ":8888"
			}
			server.ListenAndServe(&logger, address)
		},
	}
	webCmd.PersistentFlags().String("address", ":8888", "Ip and Port where the webserver is started, you can leave out the ip as a shortcut.")
	rootCmd.AddCommand(webCmd)
}
