/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wanliqun/bcx-witnode-vote-award/action"
)

var voteID string 
var filterStartDateTime, filterEndDateTime string

// submmitCmd represents the submmit command
var submmitCmd = &cobra.Command{
	Use:   "submmit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("submmit called")

		err := action.InitBCXWallet()
		if err != nil {
			fmt.Printf("bcx wallet initialized error: %v\n", err.Error())	
		} else {
			action.SubmmitVotingRecords(contractName, voteID, filterStartDateTime, filterEndDateTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(submmitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// submmitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// submmitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	submmitCmd.Flags().StringVarP(&voteID, "vote-id", "v", "", "Vote id, which is unique for your witness voting(like 1:0, 1:20 etc.)")
	submmitCmd.MarkFlagRequired("vote-id")

	submmitCmd.Flags().StringVarP(&contractName, "contract-name", "c", "", "Contract name, make sure you have already deployed it")
	submmitCmd.MarkFlagRequired("contract-name")

	submmitCmd.Flags().StringVarP(&filterStartDateTime, "filter-start-dt", "s", "0000-00-00 00:00:00", "Filter start datetime, format 'yyyy-mm-dd hh:mm:ss'")
	submmitCmd.Flags().StringVarP(&filterEndDateTime, "filter-end-dt", "e", "2099-12-31 23:59:59", "Filter end datetime, format 'yyyy-mm-dd hh:mm:ss'")
}