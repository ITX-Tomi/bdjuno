package top_accounts

import (
	"fmt"

	"github.com/forbole/bdjuno/v3/modules/bank"
	"github.com/forbole/bdjuno/v3/modules/distribution"
	"github.com/forbole/bdjuno/v3/modules/staking"
	topaccounts "github.com/forbole/bdjuno/v3/modules/top_accounts"
	modulestypes "github.com/forbole/bdjuno/v3/modules/types"
	"github.com/rs/zerolog/log"

	parsecmdtypes "github.com/forbole/juno/v3/cmd/parse/types"
	"github.com/forbole/juno/v3/types/config"
	"github.com/spf13/cobra"

	"github.com/forbole/bdjuno/v3/database"
	"github.com/forbole/bdjuno/v3/modules/auth"
)

func allCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use: "all",
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			sources, err := modulestypes.BuildSources(config.Cfg.Node, parseCtx.EncodingConfig)
			if err != nil {
				return err
			}

			// Get the database
			db := database.Cast(parseCtx.Database)

			// Build modules
			authModule := auth.NewModule(sources.AuthSource, nil, parseCtx.EncodingConfig.Marshaler, db)
			stakingModule := staking.NewModule(sources.StakingSource, nil, parseCtx.EncodingConfig.Marshaler, db)
			bankModule := bank.NewModule(nil, sources.BankSource, parseCtx.EncodingConfig.Marshaler, db)
			distiModule := distribution.NewModule(sources.DistrSource, parseCtx.EncodingConfig.Marshaler, db)
			topaccountsModule := topaccounts.NewModule(nil, nil, nil, nil, parseCtx.EncodingConfig.Marshaler, db)

			accounts, err := authModule.GetAllBaseAccounts(0)
			if err != nil {
				return fmt.Errorf("error while getting account base accounts: %s", err)
			}

			err = db.SaveAccounts(accounts)
			if err != nil {
				return err
			}

			for _, account := range accounts {
				address := account.Address

				err := bankModule.UpdateBalances([]string{address}, 0)
				if err != nil {
					log.Error().Msgf("error while refreshing account balance of account %s", address)
				}

				err = stakingModule.RefreshDelegations(0, address)
				if err != nil {
					log.Error().Msgf("error while refreshing delegations of account %s", address)
				}

				err = stakingModule.RefreshRedelegations(0, address)
				if err != nil {
					log.Error().Msgf("error while refreshing redelegations of account %s", address)
				}

				err = stakingModule.RefreshUnbondings(0, address)
				if err != nil {
					log.Error().Msgf("error while refreshing unbonding delegations of account %s", address)
				}

				err = distiModule.RefreshDelegatorRewards(0, []string{address})
				if err != nil {
					log.Error().Msgf("error while refreshing rewards of account %s", address)
				}

				err = topaccountsModule.RefreshTopAccountsSum([]string{address})
				if err != nil {
					log.Error().Msgf("error while refreshing top account sum of account %s", address)
				}
			}

			return nil
		},
	}
}
