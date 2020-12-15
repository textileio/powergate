package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client/admin"
	c "github.com/textileio/powergate/cmd/pow/common"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	adminJobsQueuedCmd.Flags().StringP("user-id", "i", "", "optional instance id filter to apply")
	adminJobsQueuedCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminJobsExecutingCmd.Flags().StringP("user-id", "i", "", "optional instance id filter to apply")
	adminJobsExecutingCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminJobsLatestFinalCmd.Flags().StringP("user-id", "i", "", "optional instance id filter to apply")
	adminJobsLatestFinalCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminJobsLatestSuccessfulCmd.Flags().StringP("user-id", "i", "", "optional instance id filter to apply")
	adminJobsLatestSuccessfulCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminJobsSummaryCmd.Flags().StringP("user-id", "i", "", "optional instance id filter to apply")
	adminJobsSummaryCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminJobsCmd.AddCommand(
		adminJobsQueuedCmd,
		adminJobsExecutingCmd,
		adminJobsLatestFinalCmd,
		adminJobsLatestSuccessfulCmd,
		adminJobsSummaryCmd,
	)
}

var adminJobsQueuedCmd = &cobra.Command{
	Use:   "queued",
	Short: "List queued storage jobs",
	Long:  `List queued storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.StorageJobs.Queued(c.AdminAuthCtx(ctx), storageJobsOpts()...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}

var adminJobsExecutingCmd = &cobra.Command{
	Use:   "executing",
	Short: "List executing storage jobs",
	Long:  `List executing storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.StorageJobs.Executing(c.AdminAuthCtx(ctx), storageJobsOpts()...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}

var adminJobsLatestFinalCmd = &cobra.Command{
	Use:   "latest-final",
	Short: "List the latest final storage jobs",
	Long:  `List the latest final storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.StorageJobs.LatestFinal(c.AdminAuthCtx(ctx), storageJobsOpts()...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}

var adminJobsLatestSuccessfulCmd = &cobra.Command{
	Use:   "latest-successful",
	Short: "List the latest successful storage jobs",
	Long:  `List the latest successful storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.StorageJobs.LatestSuccessful(c.AdminAuthCtx(ctx), storageJobsOpts()...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}

var adminJobsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Give a summary of storage jobs in all states",
	Long:  `Give a summary of storage jobs in all states`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		c.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), c.CmdTimeout)
		defer cancel()

		res, err := c.PowClient.Admin.StorageJobs.Summary(c.AdminAuthCtx(ctx), storageJobsOpts()...)
		c.CheckErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		c.CheckErr(err)

		fmt.Println(string(json))
	},
}

func storageJobsOpts() []admin.StorageJobsOption {
	var opts []admin.StorageJobsOption
	if viper.IsSet("user-id") {
		opts = append(opts, admin.WithUserID(viper.GetString("user-id")))
	}
	if viper.IsSet("cids") {
		opts = append(opts, admin.WithCids(viper.GetStringSlice("cids")...))
	}
	return opts
}
