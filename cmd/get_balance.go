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

var (
	address string
)

// getBalanceCmd represents the getBalance command
var getBalanceCmd = &cobra.Command{
	Use:   "getBalance",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		c := blockchain.NewBlockChain()
		chainState := blockchain.ChainState{c}

		wallets, err := blockchain.NewWalletManager()
		if err != nil {
			logrus.Panicf("NewTransaction GetWallet err: %v", err)
		}
		wallet := wallets.GetWallet(address)
		if wallet == nil {
			logrus.Panicf("GetBalance GetWallet not exist: %s", address)
		}

		balance, _ := chainState.FindSpendableOutputs(blockchain.HashPublicKey(wallet.PublicKey))

		logrus.Infof("Balance of %s: %d", address, balance)
	},
}

func init() {
	RootCmd.AddCommand(getBalanceCmd)

	getBalanceCmd.Flags().StringVarP(&address, "address", "a", "", "")
}
