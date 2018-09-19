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
	"github.com/profzone/chain/blockchain"
)

// getWalletsCmd represents the getWallets command
var getWalletsCmd = &cobra.Command{
	Use:   "getWallets",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		wallets, _ := blockchain.NewWalletManager()
		addresses := wallets.GetAddress()
		for _, address := range addresses {
			wallet := wallets.GetWallet(address)
			logrus.Infof("alias: %s, address: %s", wallet.Alias, address)
		}
	},
}

func init() {
	RootCmd.AddCommand(getWalletsCmd)
}
