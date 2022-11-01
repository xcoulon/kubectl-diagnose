package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xcoulon/kubectl-diagnose/pkg/diagnose"
	"github.com/xcoulon/kubectl-diagnose/pkg/logr"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Execute() {
	if err := NewDiagnoseCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewDiagnoseCmd() *cobra.Command {

	var kubeconfig string
	var namespace string
	var loglevel int

	cmd := &cobra.Command{
		Use:           "diagnose (TYPE NAME | TYPE/NAME)",
		Short:         "try to find the cause of the 503 error",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logr.New(cmd.OutOrStdout())
			logger.SetLevel(loglevel)
			kind, name, err := getResourceTypeName(args)
			if err != nil {
				return err
			}
			// look-up the kubeconfig to use
			// use the current context in kubeconfig
			cfg, config, err := newClientFromConfig(kubeconfig)
			if err != nil {
				return err
			}
			if namespace == "" {
				if namespace, _, err = config.Namespace(); err != nil {
					return err
				}
			}
			found, err := diagnose.Diagnose(logger, cfg, kind, namespace, name)
			if err != nil {
				return err
			}
			if !found {
				logger.Infof("🤷 couldn't find the culprit")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "", "", "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "(optional) the namespace scope for this CLI request")
	cmd.Flags().IntVarP(&loglevel, "loglevel", "v", 0, "log level for V logs (set to 1 or higher to display DEBUG messages)")

	return cmd
}

func getResourceTypeName(args []string) (string, string, error) {
	switch len(args) {
	case 1:
		s := strings.Split(args[0], "/")
		if len(s) != 2 {
			return "", "", fmt.Errorf("invalid resource name: %s", args[0])
		}
		return strings.ToLower(s[0]), s[1], nil
	case 2:
		if !strings.Contains(args[0], "/") && !strings.Contains(args[1], "/") {
			return strings.ToLower(args[0]), args[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid args: %v", args)
}

func newClientFromConfig(kubeconfig string) (*rest.Config, clientcmd.ClientConfig, error) {
	r, err := os.Open(locateKubeconfig(kubeconfig))
	if err != nil {
		return nil, nil, err
	}
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	config, err := clientcmd.NewClientConfigFromBytes(d)
	if err != nil {
		return nil, nil, err
	}
	cfg, err := config.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	return cfg, config, nil
}

// locateKubeconfig returns a file reader on (by order of match):
// - the --kubeconfig CLI argument if it was provided
// - the $KUBECONFIG file it the env var was set
// - the <user_home_dir>/.kube/config file
func locateKubeconfig(kubeconfig string) string {
	var path string
	if kubeconfig != "" {
		path = kubeconfig
	} else if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		path = kubeconfig
	} else {
		path = filepath.Join(homeDir(), ".kube", "config")
	}
	return path
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
