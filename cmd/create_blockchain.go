// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
	"git.profzone.net/profzone/chain/blockchain"
)

var createBlockChainAddress string

// createBlockChainCmd represents the createBlockchain command
var createBlockChainCmd = &cobra.Command{
	Use:   "createBlockChain",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if !blockchain.ValidateAddress(createBlockChainAddress) {
			logrus.Panicf("not a valid address: %s", createBlockChainAddress)
		}
		c := blockchain.InitBlockChain(createBlockChainAddress)

		chainState := blockchain.ChainState{
			BlockChain: c,
		}
		chainState.Reindex()

		logrus.Infof("chain created, reward to %s. chain state reindex complete", createBlockChainAddress)
	},
}

func init() {
	RootCmd.AddCommand(createBlockChainCmd)
	createBlockChainCmd.Flags().StringVarP(&createBlockChainAddress, "address", "a", "", "")
}
