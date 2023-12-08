package terraform

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/DavidGamba/dgtools/bt/config"
	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/icza/gox/osx"
)

// removeContainer is a function that kills and removes a Docker container.
// It takes a context.Context, a *client.Client, and a containerId string as parameters.
// It returns an error if there was a failure in killing or removing the container.
func removeContainer(ctx context.Context, cli *client.Client, containerId string) error {
	// Log a message indicating the container being killed and removed
	Logger.Printf("Killing and removing container %s\n", containerId)

	// Kill the container with SIGKILL signal
	if err := cli.ContainerKill(ctx, containerId, "SIGKILL"); err != nil {
		// Check if the error indicates that the container is not running
		if !strings.Contains(err.Error(), "is not running") {
			return fmt.Errorf("failed to kill container: %w", err)
		}
	}

	// Remove the container
	if err := cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{}); err != nil {
		// Check if the error indicates that the container was not found
		if !client.IsErrNotFound(err) {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	return nil
}

// cleanup is a function that performs cleanup operations after processing a Terraform plan.
// It removes the Docker container associated with the plan and deletes the plan JSON file.
// Parameters:
//   - ctx: The context.Context object for cancellation and timeouts.
//   - cli: The *client.Client object for interacting with the Docker API.
//   - containerId: The ID of the Docker container to be removed.
//   - planJson: The path to the Terraform plan JSON file to be deleted.
//
// Returns:
//   - error: An error if any cleanup operation fails, otherwise nil.
func cleanup(ctx context.Context, cli *client.Client, containerId string, planJson string) error {
	if err := removeContainer(ctx, cli, containerId); err != nil {
		return fmt.Errorf("failed to cleanup container: %s", err)
	}
	if err := os.Remove(planJson); err != nil {
		return fmt.Errorf("failed to TF plan json file: %s", err)
	}
	return nil
}

// visualizePlanCMD is a function that creates and returns a command option for the "visualize-plan" command.
// It takes a context.Context and a parent *getoptions.GetOpt as input parameters.
// The function retrieves the "profile" value from the parent option and creates a new command option with the name "visualize-plan".
// The command option is configured with a command function visualizePlanRun.
// It also retrieves a list of valid workspaces using the validWorkspaces function and adds a string option "ws" with the valid workspaces as valid values.
// The function returns the created command option.
func visualizePlanCMD(ctx context.Context, parent *getoptions.GetOpt) *getoptions.GetOpt {
	profile := parent.Value("profile").(string)

	cfg := config.ConfigFromContext(ctx)

	opt := parent.NewCommand("visualize-plan", "")
	opt.SetCommandFn(visualizePlanRun)

	wss, err := validWorkspaces(cfg, profile)
	if err != nil {
		Logger.Printf("WARNING: failed to list workspaces: %s\n", err)
	}
	opt.String("ws", "", opt.ValidValues(wss...), opt.Description("Workspace to use"))

	return opt
}

// visualizePlanRun is a function that visualizes the Terraform plan by running a Docker container
// and opening a web browser to display the plan in a visual format.
// It takes a context.Context, a *getoptions.GetOpt, and a []string as input parameters.
// It returns an error if any error occurs during the execution.
func visualizePlanRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	profile := opt.Value("profile").(string)
	ws := opt.Value("ws").(string)

	cfg := config.ConfigFromContext(ctx)
	Logger.Printf("cfg: %s\n", cfg.TFProfile[profile])

	ws, err := updateWSIfSelected(cfg.Config.DefaultTerraformProfile, profile, ws)
	if err != nil {
		return err
	}

	if cfg.TFProfile[profile].Workspaces.Enabled {
		if !workspaceSelected(cfg.Config.DefaultTerraformProfile, profile) {
			if ws == "" {
				return fmt.Errorf("running in workspace mode but no workspace selected or --ws given")
			}
		}
	}

	cmd := []string{cfg.TFProfile[profile].BinaryName, "show"}
	cmd = append(cmd, args...)
	// possitional arg goes at the end
	cmd = append(cmd, "-json")
	if ws == "" {
		cmd = append(cmd, ".tf.plan")
	} else {
		cmd = append(cmd, fmt.Sprintf(".tf.plan-%s", ws))
	}
	dataDir := fmt.Sprintf("TF_DATA_DIR=%s", getDataDir(cfg.Config.DefaultTerraformProfile, profile))
	Logger.Printf("export %s\n", dataDir)
	ri := run.CMD(cmd...).Ctx(ctx).Stdin().Log().Env(dataDir)
	if ws != "" {
		wsEnv := fmt.Sprintf("TF_WORKSPACE=%s", ws)
		Logger.Printf("export %s\n", wsEnv)
		ri.Env(wsEnv)
	}

	stdout, err := ri.STDOutOutput()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	plan_json_file_name := ".tf.plan.json"

	f, err := os.Create(plan_json_file_name)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(stdout)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	Logger.Printf("Starting Rover")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	containerName := "bt-visualize-plan"

	imageName := "im2nguyen/rover"

	// List all containers
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Check if the container exists
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.Contains(name, containerName) {
				Logger.Printf("Found existing container %s\n", name)
				// If the container exists, remove it
				if err := removeContainer(ctx, cli, container.ID); err != nil {
					return fmt.Errorf("failed to remove existing container: %w", err)
				}
				break
			}
		}
	}

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current dir: %w", err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Tty:   true,
		Cmd:   []string{"-planJSONPath=.tf.json"},
		ExposedPorts: nat.PortSet{
			"9000/tcp": struct{}{},
		},
	}, &container.HostConfig{
		Binds: []string{filepath.Join(cwd, plan_json_file_name) + ":/src/.tf.json"},
		PortBindings: nat.PortMap{
			"9000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "9000",
				},
			},
		},
	}, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cuctx := context.Background()
		cleanup(cuctx, cli, resp.ID, plan_json_file_name)
		os.Exit(1)
	}()

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Attach to the container
	hijackedResponse, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to container: %w", err)
	}

	// Open the default browser
	err = osx.OpenDefault("http://localhost:9000/")
	if err != nil {
		return fmt.Errorf("failed to open web browser: %w", err)
	}

	// Copy the container's output to the program's output=
	_, err = io.Copy(os.Stdout, hijackedResponse.Reader)
	if err != nil {
		return fmt.Errorf("failed to copy container's output to stdout: %w", err)
	}

	return nil
}
