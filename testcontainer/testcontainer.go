package testcontainer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go"
)

// ReaperImage - is very important to re-define because by default testcontainers lib use reaper image based on docker.io ->
// https://github.com/testcontainers/testcontainers-go/blob/6ba6e7a0e4b0046507c28e24946d595a65a96dbf/reaper.go#L24
var (
	IsSkipReaperImage = true
	ReaperImage       = "docker.io/testcontainers/ryuk:0.3.3"
)

const NoUseAuth = ""

var (
	// AuthRegistryCredStr - auth encoded string for docker registry.
	AuthRegistryCredStr = func() string { ////nolint: gochecknoglobals // this is for tests purposes.
		mustMarshalFunc := func(s interface{}) []byte {
			b, err := json.Marshal(s)
			if err != nil {
				panic(err)
			}

			return b
		}

		if registryToken := os.Getenv("REGISTRY_TOKEN"); registryToken != "" {
			return base64.URLEncoding.EncodeToString(
				mustMarshalFunc(
					&types.AuthConfig{
						RegistryToken: registryToken,
					}),
			)
		}

		return NoUseAuth
	}()
)

// PullDockerImage - pull docker image via docker pull cmd.
func PullDockerImage(ctx context.Context, imageName string) error {
	if AuthRegistryCredStr != "" {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("could not create dcoker client with custom  not pull image: %s; err: %w",
				imageName, err)
		}

		out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{RegistryAuth: AuthRegistryCredStr})
		if err != nil {
			return fmt.Errorf("could not pull image: %s; err: %w", imageName, err)
		}

		defer func() { _ = out.Close() }()
		_, _ = io.Copy(os.Stdout, out)
	}

	fmt.Println("WARN! `AuthRegistryCredStr` is empty. NOT PULL ANY IMAGES.")

	return nil
}

// BaseContainerConfig - base container request config.
type BaseContainerConfig struct {
	Name         string // Hostname, ContainerName, Network alias
	Image        string
	Port         string
	ExposedPorts []string
	Envs         map[string]string
}

// GetBaseContainerRequest - return base ContainerRequest.
func GetBaseContainerRequest(
	cfg BaseContainerConfig,
	dockerNetwork *testcontainers.DockerNetwork,
) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		ReaperImage:    ReaperImage,
		SkipReaper:     IsSkipReaperImage,
		RegistryCred:   AuthRegistryCredStr,
		Hostname:       cfg.Name,
		Image:          cfg.Image,
		ExposedPorts:   []string{cfg.Port},
		Env:            cfg.Envs,
		Networks:       []string{dockerNetwork.Name},
		NetworkAliases: map[string][]string{dockerNetwork.Name: {cfg.Name}},
	}
}
