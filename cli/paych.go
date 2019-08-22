package cli

import (
	"fmt"

	"github.com/filecoin-project/go-lotus/chain/address"
	types "github.com/filecoin-project/go-lotus/chain/types"
	"gopkg.in/urfave/cli.v2"
)

var paychCmd = &cli.Command{
	Name:  "paych",
	Usage: "Manage payment channels",
	Subcommands: []*cli.Command{
		paychCreateCmd,
		paychListCmd,
		paychVoucherCmd,
	},
}

var paychCreateCmd = &cli.Command{
	Name:  "create",
	Usage: "Create a new payment channel",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("must pass three arguments: <from> <to> <amount>")
		}

		from, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return fmt.Errorf("failed to parse from address: %s", err)
		}

		to, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to parse to address: %s", err)
		}

		amt, err := types.BigFromString(cctx.Args().Get(2))
		if err != nil {
			return fmt.Errorf("parsing amount failed: %s", err)
		}

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		addr, err := api.PaychCreate(ctx, from, to, amt)
		if err != nil {
			return err
		}

		fmt.Println(addr.String())
		return nil
	},
}

var paychListCmd = &cli.Command{
	Name:  "list",
	Usage: "List all locally registered payment channels",
	Action: func(cctx *cli.Context) error {
		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		chs, err := api.PaychList(ctx)
		if err != nil {
			return err
		}

		for _, v := range chs {
			fmt.Println(v.String())
		}
		return nil
	},
}

var paychVoucherCmd = &cli.Command{
	Name:  "voucher",
	Usage: "Interact with payment channel vouchers",
	Subcommands: []*cli.Command{
		paychVoucherCreateCmd,
		paychVoucherCheckCmd,
		paychVoucherAddCmd,
		paychVoucherListCmd,
		paychVoucherBestSpendableCmd,
		paychVoucherSubmitCmd,
	},
}

var paychVoucherCreateCmd = &cli.Command{
	Name:  "create",
	Usage: "Create a signed payment channel voucher",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "lane",
			Value: 0,
			Usage: "specify payment channel lane to use",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass two arguments: <channel> <amount>")
		}

		ch, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		amt, err := types.BigFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		lane := cctx.Int("lane")

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		sv, err := api.PaychVoucherCreate(ctx, ch, amt, uint64(lane))
		if err != nil {
			return err
		}

		enc, err := sv.EncodedString()
		if err != nil {
			return err
		}

		fmt.Println(enc)
		return nil
	},
}

var paychVoucherCheckCmd = &cli.Command{
	Name:  "check",
	Usage: "Check validity of payment channel voucher",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass payment channel address and voucher to validate")
		}

		ch, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		sv, err := types.DecodeSignedVoucher(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		if err := api.PaychVoucherCheckValid(ctx, ch, sv); err != nil {
			return err
		}

		fmt.Println("voucher is valid")
		return nil
	},
}

var paychVoucherAddCmd = &cli.Command{
	Name:  "add",
	Usage: "Add payment channel voucher to local datastore",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass payment channel address and voucher")
		}

		ch, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		sv, err := types.DecodeSignedVoucher(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		if err := api.PaychVoucherAdd(ctx, ch, sv); err != nil {
			return err
		}

		return nil
	},
}

var paychVoucherListCmd = &cli.Command{
	Name:  "list",
	Usage: "List stored vouchers for a given payment channel",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 1 {
			return fmt.Errorf("must pass payment channel address")
		}

		ch, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		vouchers, err := api.PaychVoucherList(ctx, ch)
		if err != nil {
			return err
		}

		for _, v := range vouchers {
			fmt.Printf("Lane %d, Nonce %d: %s\n", v.Lane, v.Nonce, v.Amount.String())
		}

		return nil
	},
}

var paychVoucherBestSpendableCmd = &cli.Command{
	Name:  "best-spendable",
	Usage: "Print voucher with highest value that is currently spendable",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 1 {
			return fmt.Errorf("must pass payment channel address")
		}

		ch, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		vouchers, err := api.PaychVoucherList(ctx, ch)
		if err != nil {
			return err
		}

		var best *types.SignedVoucher
		for _, v := range vouchers {
			spendable, err := api.PaychVoucherCheckSpendable(ctx, ch, v, nil, nil)
			if err != nil {
				return err
			}
			if spendable {
				if best == nil || types.BigCmp(v.Amount, best.Amount) > 0 {
					best = v
				}
			}
		}

		if best == nil {
			return fmt.Errorf("No spendable vouchers for that channel")
		}

		enc, err := best.EncodedString()
		if err != nil {
			return err
		}

		fmt.Println(enc)
		fmt.Printf("Amount: %s\n", best.Amount)
		return nil
	},
}

var paychVoucherSubmitCmd = &cli.Command{
	Name:  "submit",
	Usage: "Submit voucher to chain to update payment channel state",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must pass payment channel address and voucher")
		}

		ch, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		sv, err := types.DecodeSignedVoucher(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		api, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}

		ctx := ReqContext(cctx)

		mcid, err := api.PaychVoucherSubmit(ctx, ch, sv)
		if err != nil {
			return err
		}

		mwait, err := api.ChainWaitMsg(ctx, mcid)
		if err != nil {
			return err
		}

		if mwait.Receipt.ExitCode != 0 {
			return fmt.Errorf("message execution failed (exit code %d)", mwait.Receipt.ExitCode)
		}

		fmt.Println("channel updated succesfully")

		return nil
	},
}