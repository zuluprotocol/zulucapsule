package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"code.vegaprotocol.io/vegacapsule/generator/genesis"
	"code.vegaprotocol.io/vegacapsule/generator/tendermint"
	"code.vegaprotocol.io/vegacapsule/state"
	"github.com/spf13/cobra"
)

var templateGenesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Template genesis file for network",
	RunE: func(cmd *cobra.Command, args []string) error {
		template, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template %q: %w", templatePath, err)
		}

		networkState, err := state.LoadNetworkState(homePath)
		if err != nil {
			return fmt.Errorf("failed to load network state: %w", err)
		}

		if networkState.Empty() {
			return networkNotBootstrappedErr("template genesis")
		}

		return templateGenesis(string(template), networkState)
	},
}

func init() {
	templateGenesisCmd.PersistentFlags().BoolVar(&withMerge,
		"with-merge",
		false,
		"Defines whether the templated config should be merged with the originally initiated one",
	)
}

func templateGenesis(templateRaw string, netState *state.NetworkState) error {
	gen, err := genesis.NewGenerator(netState.Config, templateRaw)
	if err != nil {
		return err
	}

	var buff *bytes.Buffer

	if withMerge {
		buff, err = gen.ExecuteTemplate()
	} else {
		var tendermintGen *tendermint.ConfigGenerator
		tendermintGen, err = tendermint.NewConfigGenerator(netState.Config, netState.GeneratedServices.NodeSets.ToSlice())
		if err != nil {
			return err
		}

		buff, err = gen.Generate(netState.GeneratedServices.GetValidators(), tendermintGen.GenesisValidators())
	}
	if err != nil {
		return err
	}

	return outputTemplate(buff, "genesis.json")
}