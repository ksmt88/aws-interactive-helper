package cmd

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/signal"
    "time"

    "github.com/aws/aws-sdk-go-v2/service/codebuild"
    "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
    "github.com/ksmt88/aws-interactive-helper/pkg/core"
    "github.com/manifoldco/promptui"
    "github.com/spf13/cobra"
)

type CodebuildOptions struct {
    RootOptions
    Project string
}

func init() {
    var opts CodebuildOptions

    codebuildCmd := &cobra.Command{
        Use:   "codebuild",
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

            client := codebuild.NewFromConfig(cfg)

            projectName, err := selectProject(ctx, client)
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                os.Exit(1)
            }

            err = canBuild(ctx, projectName, client)
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                os.Exit(1)
            }

            startBuildInput := codebuild.StartBuildInput{
                ProjectName: &projectName,
            }
            startBuildOutput, err := client.StartBuild(ctx, &startBuildInput)
            if err != nil {
                _, _ = fmt.Fprintln(os.Stderr, err)
                os.Exit(1)
            }

            fmt.Println(startBuildOutput.Build.BuildStatus)
        },
    }

    rootCmd.AddCommand(codebuildCmd)

    codebuildCmd.Flags().StringVar(&opts.Project, "project", "", "build project")
}

func selectProject(ctx context.Context, client *codebuild.Client) (string, error) {
    listProjectOutput, err := client.ListProjects(ctx, nil)
    if err != nil {
        return "", err
    }

    if len(listProjectOutput.Projects) == 0 {
        return "", errors.New("no projects")
    }

    prompt := promptui.Select{
        Label: "Select project",
        Items: listProjectOutput.Projects,
    }

    _, projectName, err := prompt.Run()

    if err != nil {
        return "", err
    }

    return projectName, nil
}

func canBuild(ctx context.Context, projectName string, client *codebuild.Client) error {
    listBuildsForProjectInput := codebuild.ListBuildsForProjectInput{
        ProjectName: &projectName,
        SortOrder:   types.SortOrderTypeDescending,
    }
    listBuildsForProjectOutput, err := client.ListBuildsForProject(ctx, &listBuildsForProjectInput)
    if err != nil {
        return err
    }

    if len(listBuildsForProjectOutput.Ids) == 0 {
        return nil
    }

    batchGetBuildsInput := codebuild.BatchGetBuildsInput{
        Ids: []string{listBuildsForProjectOutput.Ids[0]},
    }
    batchGetBuildsOutput, err := client.BatchGetBuilds(ctx, &batchGetBuildsInput)
    if err != nil {
        return err
    }
    if batchGetBuildsOutput.Builds[0].BuildStatus == types.StatusTypeInProgress {
        return errors.New("the build is in progress")
    }

    return nil
}
