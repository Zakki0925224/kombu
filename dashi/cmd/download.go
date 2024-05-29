package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/charmbracelet/log"
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

	// if already exists, skip download
	if _, err := os.Stat(bundleImagePath); err == nil {
		return subcommands.ExitSuccess
	}

	os.MkdirAll(tmpImagePath, os.ModePerm)

	// download docker image using skopeo
	skopeo := exec.Command("skopeo", []string{"copy", "docker://" + imageName + ":" + tag, "oci:" + tmpImagePath + ":" + tag}...)
	skopeo.Stdout = os.Stdout
	skopeo.Stderr = os.Stderr

	if err := skopeo.Run(); err != nil {
		log.Error("Failed to execute command", "cmd", skopeo, "err", err)
		os.RemoveAll(tmpImagePath)
		return subcommands.ExitFailure
	}

	// convert docker image to OCI runtime bundle

	umoci := exec.Command("umoci", []string{"unpack", "--image", tmpImagePath, bundleImagePath}...)
	umoci.Stdout = os.Stdout
	umoci.Stderr = os.Stderr

	if err := umoci.Run(); err != nil {
		log.Error("Failed to execute command", "cmd", umoci, "err", err)
		os.RemoveAll(bundleImagePath)
		os.RemoveAll(tmpImagePath)
		return subcommands.ExitFailure
	}

	os.RemoveAll(tmpImagePath)

	log.Info("Downloaded OCI runtime bundle successfully")
	return subcommands.ExitSuccess
}
