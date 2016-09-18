package config

import (
	"errors"
	"fmt"
	cc "github.com/ContainX/go-springcloud/config"
	"github.com/ContainX/go-utils/encoding"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"os"
)

const (
	EnvErrorFmt = "Error creating config from env: %s"
)

// Config provides configuration information for Marathon streams and the proxy
type Config struct {
	// The URL to Marathon: ex. http://host:8080
	// Enivronment variable: BT_MARATHON_URLS
	MarthonUrls []string `json:"marthon_urls" envconfig:"marathon_urls"`

	// The basic auth username - if applicable
	// Enivronment variable: BT_USERNAME
	Username string `json:"username"`

	// The basic auth password - if applicable
	// Enivronment variable: BT_PASSWORD
	Password string `json:"password"`

	// Optional regex filter to only reload based on certain apps that match
	// ex. ^.*something.* would match all /apps/something app identifiers
	// Enivronment variable: BT_FILTER_REGEX
	FilterRegEx string `json:"filter_regex" envconfig:"filter_regex"`

	// Port to listen to HTTP requests.  Default 7777
	Port int `json:"port"`

	// Location to nginx.conf template
	Template string `json:"template"`

	/* Internal */
	Version string `json:"-"`
}

var (
	FileNotFound = errors.New("Cannot find the specified config file")
)

// AddFlags is a hook to add additional CLI Flags
func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "", "Path and filename of local configuration file. ex: config.yml")
	cmd.Flags().String("remote", "", "URI to remote config server. ex: http://server:8888")
	cmd.Flags().String("name", "beethoven", "Remote Config: The name of the app, env: CONFIG_NAME")
	cmd.Flags().String("label", "master", "Remote Config: The branch to fetch the config from, env: CONFIG_LABEL")
	cmd.Flags().String("profile", "default", "Remote Config: The profile to use, env: CONFIG_PROFILE")
}

func LoadConfigFromCommand(cmd *cobra.Command) (*Config, error) {
	remote, _ := cmd.Flags().GetString("remote")
	config, _ := cmd.Flags().GetString("config")

	if config != "" {
		return loadFromFile(config)
	}

	if remote != "" {
		name := os.Getenv("CONFIG_NAME")
		label := os.Getenv("CONFIG_LABEL")
		profile := os.Getenv("CONFIG_PROFILE")

		if name == "" {
			name, _ = cmd.Flags().GetString("name")
		}
		if label == "" {
			label, _ = cmd.Flags().GetString("label")

		}
		if profile == "" {
			profile, _ = cmd.Flags().GetString("profile")

		}
		return loadFromRemote(remote, name, label, profile)
	}

	cfg := new(Config)
	if err := envconfig.Process("bt", cfg); err != nil {
		return nil, fmt.Errorf(EnvErrorFmt, err.Error())
	}

	if len(cfg.MarthonUrls) == 0 {
		return nil, fmt.Errorf(EnvErrorFmt, "BT_MARATHON_URLS not defined")
	}
	return cfg, nil
}

// loadFromFile loads the config from a file and returns the config
func loadFromFile(configFile string) (*Config, error) {
	if configFile == "" {
		return nil, FileNotFound
	}

	encoder, err := encoding.NewEncoderFromFileExt(configFile)

	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := encoder.UnMarshalFile(configFile, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// loadFromRemote loads the config from a remote configuration server, specifically
// spring cloud config
func loadFromRemote(server, appName, label, profile string) (*Config, error) {
	client, err := cc.New(cc.Bootstrap{
		URI:     server,
		Label:   label,
		Name:    appName,
		Profile: profile,
	})

	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := client.Fetch(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

/* Config receivers */

// HttpPort is the port we serve the API with
// default 7777 if config port is undefined
func (c *Config) HttpPort() int {
	if c.Port == 0 {
		return 7777
	}
	return c.Port
}
