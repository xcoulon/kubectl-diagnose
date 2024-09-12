package cmd

import (
	"context"
	"fmt"
	"io"
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
		// fmt.Println(err)
		os.Exit(1)
	}
}

func NewDiagnoseCmd() *cobra.Command {

	var kubeconfig string
	var namespace string
	var loglevel string
	var color bool

	cmd := &cobra.Command{
		Use:           "kubectl-diagnose (TYPE NAME | TYPE/NAME)",
		Short:         "Diagnose the resource to find the cause of the 503 error",
		SilenceErrors: false,
		SilenceUsage:  true,
		Args:          cobra.RangeArgs(1, 2),
		Version:       version(),
		Run: func(cmd *cobra.Command, args []string) {
			l, err := logr.ParseLevel(loglevel)
			if err != nil {
				fmt.Fprint(cmd.ErrOrStderr(), err.Error())
			}
			logger := logr.New(cmd.OutOrStdout(), l, color)
			kind, name, err := getResourceTypeName(args)
			if err != nil {
				logger.Errorf(err.Error())
			}
			// look-up the kubeconfig to use
			// use the current context in kubeconfig
			cfg, config, err := newClientFromConfig(kubeconfig)
			if err != nil {
				logger.Errorf(err.Error())
			}
			if namespace == "" {
				if namespace, _, err = config.Namespace(); err != nil {
					logger.Errorf(err.Error())
				}
			}
			if _, err = diagnose.Diagnose(context.TODO(), logger, cfg, kind, namespace, name); err != nil {
				logger.Errorf(err.Error())
			}
		},
	}
	cmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "", "", "absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace scope for this CLI request")
	cmd.Flags().StringVar(&loglevel, "loglevel", "info", "log level to set [debug|info|error]")
	cmd.Flags().BoolVar(&color, "color", false, "colorized error messages (in red)")

	return cmd
}

func getResourceTypeName(args []string) (diagnose.ResourceKind, string, error) {
	switch len(args) {
	case 1:
		s := strings.Split(args[0], "/")
		if len(s) != 2 {
			return "", "", fmt.Errorf("missing resource name in '%s' (expected 'pod/cookie' or 'pod cookie')", args[0])
		}
		return diagnose.NewResourceKind(s[0]), s[1], nil
	case 2:
		if !strings.Contains(args[0], "/") && !strings.Contains(args[1], "/") {
			return diagnose.NewResourceKind(args[0]), args[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid args: %v", args)
}

func newClientFromConfig(kubeconfig string) (*rest.Config, clientcmd.ClientConfig, error) {
	r, err := os.Open(locateKubeconfig(kubeconfig))
	if err != nil {
		return nil, nil, err
	}
	d, err := io.ReadAll(r)
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
