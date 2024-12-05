package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/louislouislouislouis/repr8ducer/k8s"
	"github.com/louislouislouislouis/repr8ducer/ui"
)

var (
	namespace string
	podName   string
	container string
)

var cloneCmd = &cobra.Command{
	Use:   "reproduce",
	Short: "Reproduce specific pod",
	Long:  "Copy the specific docker command in your keyboard",
	Run: func(cmd *cobra.Command, args []string) {
		runCli(namespace, podName, container)
	},
}

func init() {
	cloneCmd.PersistentFlags().
		StringVarP(&namespace, "namespace", "n", "", "Namespace to work with")
	cloneCmd.PersistentFlags().
		StringVarP(&podName, "podName", "p", "", "Podname to work replicate")
	cloneCmd.PersistentFlags().
		StringVarP(&container, "container", "c", "", "Container to work replicate")

	rootCmd.AddCommand(cloneCmd)
}

func runCli(namespace, pod, container string) {
	p := tea.NewProgram(
		ui.NewModel(k8s.GetService(), ui.ModelConfig{
			Pod:       pod,
			Namespace: namespace,
			Container: container,
		}),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
