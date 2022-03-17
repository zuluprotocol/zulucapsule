package generator

import (
	"fmt"

	"code.vegaprotocol.io/vegacapsule/config"
	"code.vegaprotocol.io/vegacapsule/generator/wallet"
	"code.vegaprotocol.io/vegacapsule/types"
)

func (g *Generator) initiateNodeSet(index int, n config.NodeConfig) (*types.NodeSet, error) {
	initTNode, err := g.tendermintGen.Initiate(index, n.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate Tendermit node id %d: %w", index, err)
	}

	initVNode, err := g.vegaGen.Initiate(index, n.Mode, initTNode.HomeDir, n.NodeWalletPass, n.VegaWalletPass, n.EthereumWalletPass)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate Vega node id %d: %w", index, err)
	}

	var initDNode *types.DataNode
	// if data node binary is defined it is assumed that data-node should be deployed
	if n.DataNodeBinary != "" {
		node, err := g.dataNodeGen.Initiate(index, n.DataNodeBinary)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate Data node id %d: %w", index, err)
		}

		initDNode = node
	}

	return &types.NodeSet{
		GroupName:  n.Name,
		Index:      index,
		Name:       fmt.Sprintf("%s-nodeset-%s-%d", g.conf.Network.Name, n.Mode, index),
		Mode:       n.Mode,
		Vega:       *initVNode,
		Tendermint: *initTNode,
		DataNode:   initDNode,
	}, nil

}

func (g *Generator) initiateNodeSets() (*nodeSets, error) {
	validatorsSet := []types.NodeSet{}
	nonValidatorsSet := []types.NodeSet{}

	var index int
	for _, n := range g.conf.Network.Nodes {
		for i := 0; i < n.Count; i++ {
			nodeSet, err := g.initiateNodeSet(index, n)
			if err != nil {
				return nil, err
			}

			if n.Mode == types.NodeModeValidator {
				validatorsSet = append(validatorsSet, *nodeSet)
			} else {
				nonValidatorsSet = append(nonValidatorsSet, *nodeSet)
			}

			index++
		}
	}

	return &nodeSets{
		validators:    validatorsSet,
		nonValidators: nonValidatorsSet,
	}, nil
}

func (g *Generator) initAndConfigureFaucet(conf *config.FaucetConfig) (*types.Faucet, error) {
	initFaucet, err := g.faucetGen.InitiateAndConfigure(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to initate faucet: %w", err)
	}

	return initFaucet, nil
}

func (g *Generator) initAndConfigureWallet(conf *config.WalletConfig, validatorsSet []types.NodeSet) (*types.Wallet, error) {
	walletConfTemplate, err := wallet.NewConfigTemplate(g.conf.Network.Wallet.Template)
	if err != nil {
		return nil, err
	}

	initWallet, err := g.walletGen.InitiateWithNetworkConfig(g.conf.Network.Wallet, validatorsSet, walletConfTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to initate wallet: %w", err)
	}

	return initWallet, nil
}