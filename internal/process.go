package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	sysmount "github.com/moby/sys/mount"
)

// Struct to hold the mountpoint configuration passed via container environment variables
type MountpointConfig struct {

	// Activation flag. Omit if not present
	ActivationFlag string `json:"activation_flag,omitempty"`

	// Container mountpoint. Omit if not present
	ContainerMountPoint string `json:"container_mountpoint,omitempty"`

	// S3 bucket name. Must be present
	S3BucketName string `json:"s3_bucket_name"`

	// S3 endpoint url. Omit if not present
	S3EndpointUrl string `json:"s3_endpoint_url,omitempty"`

	// S3 prefix. Omit if not present
	S3Prefix string `json:"s3_prefix,omitempty"`
}

// Execute process using syscall.Exec
// The mountpoint program path will be in hookConfig.ProgramPath
// The host mount point will be in hookConfig.HostMountPoint
// The container mount point will be in hookConfig.ContainerMountPoint
// Also use the environment variables from the containerConfig.Process.Env to execute the process

func ExecuteMountpointProcess(env []string, mountpointConfig MountpointConfig, hookConfig Config) error {

	// Get the S3 bucket name from mountpointConfig.S3BucketName
	// If it is empty return error
	if mountpointConfig.S3BucketName == "" {
		log.Printf("S3 bucket name is empty\n")
		return fmt.Errorf("S3 bucket name is empty")
	}

	// Create the host mount point directory path
	err := os.MkdirAll(hookConfig.HostMountPoint, 0750)
	if err != nil {
		log.Printf("unable to create host mount point directory %s\n", err)
		return err
	}

	log.Printf("Executing program %s\n", hookConfig.ProgramPath)

	// Build the arguments for the process
	// The arguments will be the host mount point and other required
	// Var to hold optional arguments
	var optionalArguments []string

	// If S3_ENDPOINT_URL is present in mountpointConfig, add it to the arguments
	if mountpointConfig.S3EndpointUrl != "" {
		optionalArguments = append(optionalArguments, "--endpoint-url", mountpointConfig.S3EndpointUrl)
	}

	// If S3_PREFIX is present in mountpointConfig, add it to the arguments
	if mountpointConfig.S3Prefix != "" {
		optionalArguments = append(optionalArguments, "--prefix", mountpointConfig.S3Prefix)
	}

	arguments := []string{
		mountpointConfig.S3BucketName,
		hookConfig.HostMountPoint,
	}

	// Append optional arguments to arguments
	arguments = append(arguments, optionalArguments...)

	// Create a new command with the program path and arguments
	cmd := exec.Command(hookConfig.ProgramPath, arguments...) // #nosec G204

	// Set the environment variables for the command
	cmd.Env = env

	// Run the command
	err = cmd.Run()
	if err != nil {
		log.Printf("unable to execute process %s\n", err)
		return err
	}

	return nil
}

// Bind mount src to dst
// The src will be the host mount point and dst will be the container mount point

func BindMount(srcMountPoint string, dstMountPoint string) error {

	log.Printf("Bind mounting host mount point %s to container mount point %s\n",
		srcMountPoint, dstMountPoint)

	// Create the dst mount point directory path
	err := os.MkdirAll(dstMountPoint, 0750)
	if err != nil {
		log.Printf("create container mount point directory returned err: %s\n", err)
		return err
	}

	// Bind mount the host mount point to container mount point
	err = sysmount.Mount(srcMountPoint, dstMountPoint, "none", "bind,rw")
	if err != nil {
		log.Printf("bind mount srcMountPoint (%s) dstMountPoint (%s) returned err: %s\n", srcMountPoint, dstMountPoint, err)
		return err
	}

	return nil
}

// Process containerConfig.Process.Env and update the struct MountpointConfig

func ProcessEnv(env []string, hookConfig Config) (mountpointConfig MountpointConfig) {
	// If env is nil, return empty MountpointConfig
	if env == nil {
		return MountpointConfig{}
	}

	for _, envVar := range env {
		parts := strings.Split(envVar, "=")
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		switch key {
		case hookConfig.ActivationFlag:
			mountpointConfig.ActivationFlag = value
		case "CONTAINER_MOUNTPOINT":
			mountpointConfig.ContainerMountPoint = value
		case "S3_BUCKET":
			mountpointConfig.S3BucketName = value
		case "S3_PREFIX":
			mountpointConfig.S3Prefix = value
		case "S3_ENDPOINT_URL":
			mountpointConfig.S3EndpointUrl = value
		}
	}
	return mountpointConfig
}
