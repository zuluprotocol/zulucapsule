package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"code.vegaprotocol.io/vegacapsule/ethereum"
	"code.vegaprotocol.io/vegacapsule/state"
	"code.vegaprotocol.io/vegacapsule/types"
	"code.vegaprotocol.io/vegacapsule/utils"
	"github.com/spf13/cobra"
)

var ethereumCmd = &cobra.Command{
	Use:   "ethereum",
	Short: "Support interactions with ethereum network",
}

// Flags
var (
	ethereumAddress string
	ethereumChainID int

	ethereumWaitTimeoutSeconds uint
)

const (
	defaultEthereumAddress = "ws://127.0.0.1:8545"
	defaultEthereumChainID = 1440

	defaultEthreumWaitTimeout = 60
)

func init() {
	ethereumWaitCmd.Flags().StringVar(&ethereumAddress,
		"eth-address",
		defaultEthereumAddress,
		"Specify the ethereum network address",
	)
	ethereumWaitCmd.Flags().IntVar(&ethereumChainID,
		"eth-chain-id",
		defaultEthereumChainID,
		"Specify the ethereum chain ID",
	)

	ethereumMultisigCmd.Flags().StringVar(&ethereumAddress,
		"eth-address",
		defaultEthereumAddress,
		"Specify the ethereum network address",
	)
	ethereumMultisigCmd.Flags().IntVar(&ethereumChainID,
		"eth-chain-id",
		defaultEthereumChainID,
		"Specify the ethereum chain ID",
	)

	ethereumCmd.AddCommand(ethereumMultisigCmd)
	ethereumMultisigCmd.AddCommand(ethereumMultisigSetupCmd)

	ethereumCmd.PersistentFlags().UintVar(&ethereumWaitTimeoutSeconds,
		"timeout",
		defaultEthreumWaitTimeout,
		"Specify the number of second to wait",
	)

	ethereumCmd.AddCommand(ethereumWaitCmd)
	ethereumCmd.AddCommand(ethereumAssetCmd)
}

var ethereumMultisigCmd = &cobra.Command{
	Use:   "multisig",
	Short: "Manages multisig smartcontract",
}

var ethereumWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Waits for the ethereum network to be available",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ethereumWaitTimeoutSeconds)*time.Second)
		defer cancel()

		return ethereum.WaitForNetwork(ctx, ethereumChainID, ethereumAddress)
	},
}

var ethereumMultisigSetupCmd = &cobra.Command{
	Use:   "init",
	Short: "Setups the multisig smart contract",
	Long:  `Adds all validators to the multisig smart contract`,
	RunE: func(cmd *cobra.Command, args []string) error {
		netState, err := state.LoadNetworkState(homePath)
		if err != nil {
			return err
		}

		if netState.Empty() {
			return networkNotBootstrappedErr("ethereum multisig init")
		}

		if !netState.Running() {
			return networkNotRunningErr("ethereum multisig init")
		}

		smartcontracts, err := netState.Config.SmartContractsInfo()
		if err != nil {
			return fmt.Errorf("failed getting smart contract informations: %w", err)
		}

		chainID, err := strconv.Atoi(netState.Config.Network.Ethereum.ChainID)
		if err != nil {
			return err
		}

		ctx := context.Background()
		client, err := ethereum.NewEthereumMultisigClient(ctx, ethereum.EthereumMultisigClientParameters{
			VegaBinary: *netState.Config.VegaBinary,
			VegaHome:   utils.VegaNodeHomePath(homePath, 0),

			ChainID:            chainID,
			EthereumAddress:    netState.Config.Network.Ethereum.Endpoint,
			SmartcontractsInfo: *smartcontracts,
		})
		if err != nil {
			return fmt.Errorf("failed to create ethereum client: %w", err)
		}

		validatorsKeyPairs := getSigners(netState.GeneratedServices.ListValidators())
		return client.InitMultisig(ctx, *smartcontracts, validatorsKeyPairs)
	},
}

func getSigners(nodes []types.VegaNodeOutput) []ethereum.Signer {
	result := make([]ethereum.Signer, len(nodes))

	for idx, node := range nodes {
		result[idx] = ethereum.Signer{
			HomeAddress:        node.VegaNode.HomeDir,
			WalletPassFilePath: node.NodeWalletPassFilePath,
			ClefRPCAddress:     node.VegaNode.NodeWalletInfo.EthereumClefRPCAddress,
			KeyPair: ethereum.KeyPair{
				Address:    node.VegaNode.NodeWalletInfo.EthereumAddress,
				PrivateKey: node.VegaNode.NodeWalletInfo.EthereumPrivateKey,
			},
		}
	}

	return result
}
