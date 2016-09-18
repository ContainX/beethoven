package main

import (
	"fmt"
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/depcon/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	Usage = `
Beethoven (Mesos/Marathon HTTP based Proxy)

== Version: %s - Built: %s ==
`
	Example = `
   Environment   : BT_MARATHON_URLS=http://host:8080,http://host2 beethoven serve
   Local Config  : beethoven serve --config filepath
   Remote Config : beethoven serve --remote http://confighost --name myapp --profile prod
`
)

var (
	// Added via ldflags
	version = "-"
	built   = ""

	// Root command is the parent to all other commands
	rootCmd = &cobra.Command{
		Use:     "beethoven [config-file | remote-server-url]",
		Short:   "Mesos/Marathon HTTP based Proxy",
		Long:    fmt.Sprintf(Usage, version, built),
		Example: Example,
	}

	serveCmd = &cobra.Command{
		Use:     "serve",
		Short:   "Start serving traffic",
		Run:     serve,
		Example: Example,
	}

	log = logger.GetLogger("beethoven")
)

func serve(cmd *cobra.Command, args []string) {
	config, err := config.LoadConfigFromCommand(cmd)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(config)
}

func main() {
	rootCmd.AddCommand(serveCmd)
	config.AddFlags(serveCmd)
	rootCmd.Execute()
}
