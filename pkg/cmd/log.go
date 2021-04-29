package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "time"

    "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
    "github.com/ksmt88/aws-interactive-helper/pkg/core"
    "github.com/manifoldco/promptui"
    "github.com/spf13/cobra"
)

type LogOptions struct {
    RootOptions
    LogGroupNamePrefix string
    Filter             string
}

func init() {
    var opts LogOptions

    logCmd := &cobra.Command{
        Use:   "log",
        Short: "",
        Long:  "",
        Run: func(cmd *cobra.Command, args []string) {
            ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
            defer stop()
            ctx, cancel := context.WithTimeout(ctx, time.Second*5)
            defer cancel()

            region, err := rootCmd.Flags().GetString("region")
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                os.Exit(1)
            }

            profile, err := rootCmd.Flags().GetString("profile")
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                os.Exit(1)
            }

            opts.Region = region
            opts.Profile = profile

            cfg, err := core.NewConfig(ctx, region, profile)
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                os.Exit(1)
            }

            client := cloudwatchlogs.NewFromConfig(cfg)

            logGroupsInput := cloudwatchlogs.DescribeLogGroupsInput{
                Limit:              nil,
                LogGroupNamePrefix: nil,
            }
            logGroupsOutput, err := client.DescribeLogGroups(ctx, &logGroupsInput)
            if err != nil {
                fmt.Println(err)
            }

            var items []string
            for _, group := range logGroupsOutput.LogGroups {
                items = append(items, *group.LogGroupName)
            }
            promptSelect := promptui.Select{
                Label: "Select LogGroup",
                Items: items,
            }

            _, logGroup, err := promptSelect.Run()

            if err != nil {
                fmt.Printf("Prompt failed %v\n", err)
                return
            }

            promptInput := promptui.Prompt{
                Label: "Input search text",
            }

            filter, err := promptInput.Run()
            if err != nil {
                fmt.Printf("Prompt failed %v\n", err)
                return
            }

            filterLogEventsInput := cloudwatchlogs.FilterLogEventsInput{
                // EndTime:    nil,
                FilterPattern: &filter,
                LogGroupName:  &logGroup,
                // StartTime:  nil,
            }
            filterLogEventsOutput, err := client.FilterLogEvents(ctx, &filterLogEventsInput)
            if err != nil {
                fmt.Println(err)
            }

            for _, event := range filterLogEventsOutput.Events {
                fmt.Println(*event.Message, *event.Timestamp)
            }
        },
    }

    rootCmd.AddCommand(logCmd)

    logCmd.Flags().StringVar(&opts.LogGroupNamePrefix, "prefix", "", "log group name prefix")
    logCmd.Flags().StringVar(&opts.Filter, "filter", "", "search text")
}
