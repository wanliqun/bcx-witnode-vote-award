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
	"time"

	"github.com/spf13/cobra"
	"github.com/wanliqun/bcx-witnode-vote-award/action"
)

var contractName, contractPath string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy called")

		err := action.InitBCXWallet()
		if err != nil {
			fmt.Printf("bcx wallet initialized error: %v\n", err.Error())	
		} else {
			action.DeploySmartContract(contractName, contractPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	t := time.Now()
	defaultContractName := fmt.Sprintf("contract.witvote-reward-%v", t.Unix())
	deployCmd.Flags().StringVarP(&contractName, "contract-name", "n", defaultContractName, "Deploy contract name")

	deployCmd.Flags().StringVarP(&contractPath, "contract-path", "p", "./asset/contract/witness_vote_reward.lua", "Deploy lua contract file path")
}