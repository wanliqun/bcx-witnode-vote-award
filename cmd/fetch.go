package cmd

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"
	"github.com/wanliqun/bcx-witnode-vote-award/action"
)

var startBlock, endBlock int64

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fetch called")
		fmt.Printf("%v %v\n", startBlock, endBlock)
		action.FetchBlocks(startBlock, endBlock)
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	fetchCmd.Flags().Int64VarP(&startBlock, "start", "s", 0, "Start block number")
	fetchCmd.Flags().Int64VarP(&endBlock, "end", "e", math.MaxInt64, "End block number")
}
