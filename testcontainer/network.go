package testcontainer

import (
	"context"
)

// NewDockerNetwork - create new docker network.
func NewDockerNetwork(ctx context.Context, name string) (Network, error) {
	if err := PullDockerImage(ctx, ReaperImage); err != nil {
		return nil, err
	}

	return GenericNetwork(
		ctx, GenericNetworkRequest{
			NetworkRequest: NetworkRequest{
				Name:        name,
				ReaperImage: ReaperImage,
				SkipReaper:  IsSkipReaperImage,
			},
		},
	)
}
