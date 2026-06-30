package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/hyperledger-firefly/tezosconnect/internal/msgs"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var shortened = false
var output = "json"

var BuildDate string            // set by go-releaser
var BuildCommit string          // set by go-releaser
var BuildVersionOverride string // set by go-releaser

type Info struct {
	Version string `json:"Version,omitempty" yaml:"Version,omitempty"`
	Commit  string `json:"Commit,omitempty" yaml:"Commit,omitempty"`
	Date    string `json:"Date,omitempty" yaml:"Date,omitempty"`
	License string `json:"License,omitempty" yaml:"License,omitempty"`
}

func setBuildInfo(info *Info, buildInfo *debug.BuildInfo, ok bool) {
	if ok {
		info.Version = buildInfo.Main.Version
	}
}

func versionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the version info",
		Long:  "",
		RunE: func(_ *cobra.Command, _ []string) error {
			info := &Info{
				Version: BuildVersionOverride,
				Date:    BuildDate,
				Commit:  BuildCommit,
				License: "Apache-2.0",
			}

			// When we are using go install, we will get version information from Go
			// When we're in go-releaser in a Github action, we will have the version passed in explicitly
			if info.Version == "" {
				buildInfo, ok := debug.ReadBuildInfo()
				setBuildInfo(info, buildInfo, ok)
			}

			if shortened {
				fmt.Printf("Version: %s", info.Version)
			} else {
				var (
					bytes []byte
					err   error
				)

				switch output {
				case "json":
					bytes, err = json.MarshalIndent(info, "", "  ")
				case "yaml":
					bytes, err = yaml.Marshal(info)
				default:
					err = i18n.NewError(context.Background(), msgs.MsgInvalidOutputType, output)
				}

				if err != nil {
					return err
				}

				fmt.Printf("Build info: %s", string(bytes))
			}

			return nil
		},
	}

	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "print only the version")
	versionCmd.Flags().StringVarP(&output, "output", "o", "json", "output format (\"yaml\"|\"json\")")
	return versionCmd
}
