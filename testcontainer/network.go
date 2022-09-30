package testcontainer

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// NewDockerNetwork - create new docker network.
func NewDockerNetwork(ctx context.Context, name string) (testcontainers.Network, error) {
	if err := PullDockerImage(ctx, ReaperImage); err != nil {
		return nil, err
	}

	return testcontainers.GenericNetwork(
		ctx, testcontainers.GenericNetworkRequest{
			NetworkRequest: testcontainers.NetworkRequest{
				Name:        name,
				ReaperImage: ReaperImage,
				SkipReaper:  IsSkipReaperImage,
			},
		},
	)
}
