package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type Download struct{}

func (t *Download) Name() string { return "download" }
func (t *Download) Synopsis() string {
	return "download docker image and convert to OCI runtime bundle"
}
func (t *Download) Usage() string            { return "download <docker image name> <tag>: " + t.Synopsis() }
func (t *Download) SetFlags(f *flag.FlagSet) {}
func (t *Download) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 2 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	imageName := args[0]
	tag := args[1]

	bundleImagePath := internal.OCI_RUNTIME_BUNDLES_PATH + "/" + imageName + "-" + tag
	tmpImagePath := bundleImagePath + "-tmp"
	os.MkdirAll(tmpImagePath, os.ModePerm)

	// download docker image using skopeo
	skopeo := exec.Command("skopeo", []string{"copy", "docker://" + imageName + ":" + tag, "oci:" + tmpImagePath + ":" + tag}...)
	skopeo.Stdout = os.Stdout
	skopeo.Stderr = os.Stderr

	if err := skopeo.Run(); err != nil {
		fmt.Printf("Failed to execute command: %v: %s\n", skopeo, err)
		os.RemoveAll(tmpImagePath)
		return subcommands.ExitFailure
	}

	// convert docker image to OCI runtime bundle
	umoci := exec.Command("umoci", []string{"unpack", "--image", tmpImagePath, bundleImagePath}...)
	umoci.Stdout = os.Stdout
	umoci.Stderr = os.Stderr

	if err := umoci.Run(); err != nil {
		fmt.Printf("Failed to execute command: %v: %s\n", umoci, err)
		os.RemoveAll(bundleImagePath)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
