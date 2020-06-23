package cmd

import (
	"errors"
	"io"
	"kube-recreate/pkg/k8s"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ingressCmd struct {
	out      io.Writer
	ns       string
	reporter *Reporter
}

func NewIngressCommand(streams genericclioptions.IOStreams) *cobra.Command {
	rCmd := &ingressCmd{
		out:      streams.Out,
		reporter: NewReporter(streams.Out),
	}

	cmd := &cobra.Command{
		Use:          "ingress",
		Short:        "Deletes and recreates all ingress resources",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("this command does not accept arguments")
			}
			rCmd.ns = getNamespace(genericclioptions.NewConfigFlags(true), c)

			return rCmd.run()
		},
	}

	return cmd
}

func (ir *ingressCmd) run() error {
	client, err := k8s.NewK8sClient()
	if err != nil {
		return err
	}

	l, err := client.LsIngress(ir.ns)

	for _, ingress := range l {
		client.DeleteIngress(&ingress)
		ir.reporter.Append(ingress.Name, "Ingress", "DELETED", ingress.CreationTimestamp.String())

	}

	ir.reporter.AddSeperator()

	for _, ingress := range l {
		ingress.ResourceVersion = ""

		i, err := client.CreateIngress(&ingress)
		if err != nil {
			ir.reporter.Append(ingress.Name, "Ingress", "FAILED", ingress.CreationTimestamp.String())
		}

		ir.reporter.Append(i.Name, "Ingress", "CREATED", i.CreationTimestamp.String())
	}

	ir.reporter.PrintReport()
	return nil
}