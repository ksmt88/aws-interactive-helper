package core

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
)

func NewConfig(ctx context.Context, region string, profile string) (aws.Config, error) {
    cfg, err := config.LoadDefaultConfig(
        ctx,
        config.WithRegion(region),
        config.WithSharedConfigProfile(profile),
    )
    if err != nil {
        return aws.Config{}, err
    }

    return cfg, nil
}
