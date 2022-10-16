package testcontainer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Type Aliasing for reduce code line size. Remove external deps from other places too.
type (
	Container        = testcontainers.Container
	ContainerRequest = testcontainers.ContainerRequest
	ContainerFile    = testcontainers.ContainerFile

	Network        = testcontainers.Network
	DockerNetwork  = testcontainers.DockerNetwork
	NetworkRequest = testcontainers.NetworkRequest

	GenericContainerRequest = testcontainers.GenericContainerRequest
	GenericNetworkRequest   = testcontainers.GenericNetworkRequest
)

// Func aliasing - for reduce code line size. Remove external deps from other places too.
var (
	GenericContainer = testcontainers.GenericContainer
	GenericNetwork   = testcontainers.GenericNetwork

	Mounts    = testcontainers.Mounts
	BindMount = testcontainers.BindMount
)

// BaseContainerConfig - base container request config.
type BaseContainerConfig struct {
	Name               string // Hostname, ContainerName, Network alias
	Image              string
	Port               string
	Files              []ContainerFile
	Binds              []string
	Envs               map[string]string
	Cmd                []string
	ExposedPorts       []string
	AutoRemove         bool
	WaitingForStrategy wait.Strategy
}

// IsSkipReaperImage - is skip usage Reaper Image.
const IsSkipReaperImage = false // not skip usage reaper image on CI

// ReaperImage - is very important to re-define because by default testcontainers lib use reaper image based on docker.io ->
// https://github.com/testcontainers/testcontainers-go/blob/6ba6e7a0e4b0046507c28e24946d595a65a96dbf/reaper.go#L24
const ReaperImage = "" // use standard Reaper Image.

// NoUseAuth - not used any docker auth.
const NoUseAuth = ""

// RegistryTokenEnv - env with docker registry token from dp.
const RegistryTokenEnv = "" // not use any docker registry token for docker demon.

var (
	once                sync.Once
	authRegistryCredStr string
)

type (
	Terminated interface {
		Terminate(ctx context.Context) error
	}
)

func TerminateIfNotNil(ctx context.Context, container Terminated) error {
	if container == nil {
		return nil
	}

	return container.Terminate(ctx)
}

// GetAuthRegistryCredStr - return auth encoded string for docker registry.
func GetAuthRegistryCredStr() string {
	registryToken := os.Getenv(RegistryTokenEnv)
	if registryToken == "" {
		return NoUseAuth
	}

	once.Do(func() {
		authRegistryCredStr = getAuthRegistryCredStr(registryToken)
	})

	return authRegistryCredStr
}

// getAuthRegistryCredStr - convert registry token to auth registry cred str in base64 format.
func getAuthRegistryCredStr(registryToken string) string {
	mustMarshalFunc := func(s interface{}) []byte {
		b, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}

		return b
	}

	return base64.URLEncoding.EncodeToString(
		mustMarshalFunc(
			&types.AuthConfig{
				RegistryToken: registryToken,
			}))
}

// PullDockerImage - pull docker image via docker client.
func PullDockerImage(ctx context.Context, imageName string) error {
	if GetAuthRegistryCredStr() == "" {
		fmt.Println("WARN! `AuthRegistryCredStr` is empty. NOT FORCE PULL ANY IMAGES.")

		return nil
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("could not create docker client with custom  not pull image: %s; err: %w",
			imageName, err)
	}

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{RegistryAuth: GetAuthRegistryCredStr()})
	if err != nil {
		return fmt.Errorf("could not pull image: %s; err: %w", imageName, err)
	}

	defer func() { _ = out.Close() }()
	_, _ = io.Copy(os.Stdout, out)

	return nil
}

// NewGenericContainer - create new generic container with BaseContainerConfig and DockerNetwork.
func NewGenericContainer(ctx context.Context, cfg BaseContainerConfig, dockerNetwork *DockerNetwork) (Container, error) {
	return GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: GetBaseContainerRequest(cfg, dockerNetwork),
	})
}

// GetBaseContainerRequest - return base ContainerRequest.
func GetBaseContainerRequest(
	cfg BaseContainerConfig,
	dockerNetwork *DockerNetwork,
) ContainerRequest {
	return ContainerRequest{
		Name:           cfg.Name,
		ReaperImage:    ReaperImage,
		SkipReaper:     IsSkipReaperImage,
		RegistryCred:   GetAuthRegistryCredStr(),
		Hostname:       cfg.Name,
		Image:          cfg.Image,
		ExposedPorts:   cfg.ExposedPorts,
		Env:            cfg.Envs,
		Files:          cfg.Files,
		Binds:          cfg.Binds,
		Cmd:            cfg.Cmd,
		Networks:       []string{dockerNetwork.Name},
		NetworkAliases: map[string][]string{dockerNetwork.Name: {cfg.Name}},
		AutoRemove:     cfg.AutoRemove,
		WaitingFor:     cfg.WaitingForStrategy,
	}
}
