package main

import (
	"fmt"
	"github.com/ContainX/beethoven/config"
	"github.com/ContainX/beethoven/proxy"
	"github.com/ContainX/depcon/pkg/logger"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"os"
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
	/* LDFlags */
	version = "-"
	built   = ""

	/* CLI commands */

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

	// Logger
	log    = logger.GetLogger("beethoven")
	format = logging.MustStringFormatter("%{time:2006-01-02 15:04:05} %{level:.9s} [%{module}]: %{message}")
)

func serve(cmd *cobra.Command, args []string) {
	config, err := config.LoadConfigFromCommand(cmd)
	if err != nil {
		log.Fatal(err.Error())
	}
	config.Version = version

	proxy.New(config).Serve()

}

func main() {
	setupLogging()
	rootCmd.AddCommand(serveCmd)
	config.AddFlags(serveCmd)
	rootCmd.Execute()
}

func setupLogging() {
	if os.Getenv("DOCKER_ENV") != "" {
		backend := logging.NewLogBackend(os.Stderr, "", 0)
		backendFmt := logging.NewBackendFormatter(backend, format)
		logging.SetBackend(backendFmt)
	}
}
