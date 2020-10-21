package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	adminCmd.PersistentFlags().String("admin-token", "", "admin auth token")

	adminQueuedStorageJobsCmd.Flags().StringP("instance-id", "i", "", "optional instance id filter to apply")
	adminQueuedStorageJobsCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminExecutingStorageJobsCmd.Flags().StringP("instance-id", "i", "", "optional instance id filter to apply")
	adminExecutingStorageJobsCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminLatestFinalStorageJobsCmd.Flags().StringP("instance-id", "i", "", "optional instance id filter to apply")
	adminLatestFinalStorageJobsCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminLatestSuccessfulStorageJobsCmd.Flags().StringP("instance-id", "i", "", "optional instance id filter to apply")
	adminLatestSuccessfulStorageJobsCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	adminStorageJobsSummaryCmd.Flags().StringP("instance-id", "i", "", "optional instance id filter to apply")
	adminStorageJobsSummaryCmd.Flags().StringSliceP("cids", "c", nil, "optional cids filter to apply")

	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(
		adminCreateInstanceCmd,
		adminInstancesCmd,
		adminQueuedStorageJobsCmd,
		adminExecutingStorageJobsCmd,
		adminLatestFinalStorageJobsCmd,
		adminLatestSuccessfulStorageJobsCmd,
		adminStorageJobsSummaryCmd,
	)
}

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Provides admin commands",
	Long:  `Provides admin commands`,
}

var adminCreateInstanceCmd = &cobra.Command{
	Use:   "create-profile",
	Short: "Create a Powergate storage profile.",
	Long:  `Create a Powergate storage profile.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.CreateStorageProfile(adminAuthCtx(ctx))
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminInstancesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List all Powergate storage profiles.",
	Long:  `List all Powergate storage profiles.`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.ListStorageProfiles(adminAuthCtx(ctx))
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminQueuedStorageJobsCmd = &cobra.Command{
	Use:   "queued",
	Short: "List queued storage jobs",
	Long:  `List queued storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.QueuedStorageJobs(
			adminAuthCtx(ctx),
			viper.GetString("instance-id"),
			viper.GetStringSlice("cids")...,
		)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminExecutingStorageJobsCmd = &cobra.Command{
	Use:   "executing",
	Short: "List executing storage jobs",
	Long:  `List executing storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.ExecutingStorageJobs(
			adminAuthCtx(ctx),
			viper.GetString("instance-id"),
			viper.GetStringSlice("cids")...,
		)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminLatestFinalStorageJobsCmd = &cobra.Command{
	Use:   "latest-final",
	Short: "List the latest final storage jobs",
	Long:  `List the latest final storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.LatestFinalStorageJobs(
			adminAuthCtx(ctx),
			viper.GetString("instance-id"),
			viper.GetStringSlice("cids")...,
		)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminLatestSuccessfulStorageJobsCmd = &cobra.Command{
	Use:   "latest-successful",
	Short: "List the latest successful storage jobs",
	Long:  `List the latest successful storage jobs`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.LatestSuccessfulStorageJobs(
			adminAuthCtx(ctx),
			viper.GetString("instance-id"),
			viper.GetStringSlice("cids")...,
		)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}

var adminStorageJobsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Give a summary of storage jobs in all states",
	Long:  `Give a summary of storage jobs in all states`,
	Args:  cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		checkErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()

		res, err := fcClient.Admin.StorageJobsSummary(
			adminAuthCtx(ctx),
			viper.GetString("instance-id"),
			viper.GetStringSlice("cids")...,
		)
		checkErr(err)

		json, err := protojson.MarshalOptions{Multiline: true, Indent: "  ", EmitUnpopulated: true}.Marshal(res)
		checkErr(err)

		fmt.Println(string(json))
	},
}
