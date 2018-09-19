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
	"fmt"

	"github.com/spf13/cobra"
	"strconv"
	"git.profzone.net/profzone/chain/blockchain"
)

// browseChainCmd represents the browseChain command
var browseChainCmd = &cobra.Command{
	Use:   "browseChain",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		c := blockchain.NewBlockChain()
		it := c.Iterator()

		for {
			block := it.Next()

			if block == nil {
				break
			}
			fmt.Printf("Prev hash: %x\n", block.Header.PrevBlockHash)
			fmt.Printf("Hash: %x\n", block.Header.Hash)
			fmt.Printf("Height: %d\n", block.Header.Height)

			trans := blockchain.DeserializeTransactionContainer(block.Body.Data)
			for _, tran := range trans {
				fmt.Printf("\tTran ID: %x\n", tran.ID)
				for index, input := range tran.Inputs {
					fmt.Printf("\t\t-------------------------------Input %d--------------------------------\n", index)
					fmt.Printf("\t\tTran ID:\t%x\n", input.TransactionID)
					fmt.Printf("\t\tOutIndex:\t%d\n", input.OutputIndex)
				}
				fmt.Println()

				for index, output := range tran.Outputs {
					fmt.Printf("\t\t-------------------------------Output %d-------------------------------\n", index)
					fmt.Printf("\t\tAmount: %d\n", output.Amount)
				}
				fmt.Println()
			}

			pow := blockchain.NewPOW(block)
			fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
			fmt.Println()

			if block.Header.PrevBlockHash == nil || len(block.Header.PrevBlockHash) == 0 {
				break
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(browseChainCmd)
}
