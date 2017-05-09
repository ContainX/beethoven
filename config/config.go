package config

import (
	"errors"
	"fmt"
	cc "github.com/ContainX/go-springcloud/config"
	"github.com/ContainX/go-utils/encoding"
	"github.com/ContainX/go-utils/logger"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

const (
	EnvErrorFmt                            = "Error creating config from env: %s"
	DefaultNginxTemplatePath               = "/etc/nginx/nginx.template"
	DefaultNginxConfPath                   = "/etc/nginx/nginx.conf"
	MarathonScheduler        SchedulerType = 1
	SwarmScheduler           SchedulerType = 2
)

type SchedulerType int

var log = logger.GetLogger("beethoven.config")

// Config provides configuration information for Marathon streams and the proxy
type Config struct {
	// Scheduler type to use (0 for Marathon, 1 for Swarm)
	// Only applicable if both Swarm and Marathon are configured
	SchedulerType SchedulerType `json:"scheduler_type"`

	// Docker/Swarm configuration
	Swarm *SwarmConfig `json:"swarm"`

	// Marathon configuration options
	Marathon *MarathonConfig `json:"marathon"`

	// Deprecated - Please use Marathon
	MarthonUrls []string `json:"marthon_urls" envconfig:"-"`

	// Deprecated - Use Marathon.Username
	Username string `json:"username" envconfig:"-"`

	// Deprecated - Use Marathon.Password
	Password string `json:"password" envconfig:"-"`

	// Optional regex filter to only reload based on certain apps that match
	// ex. ^.*something.* would match all /apps/something app identifiers
	// Enivronment variable: BT_FILTER_REGEX
	FilterRegExStr string `json:"filter_regex" envconfig:"filter_regex"`

	// Resolved Filter regex
	filterRegEx *regexp.Regexp

	// Port to listen to HTTP requests.  Default 7777
	Port int `json:"port"`

	// Scheme we are listening to (http | https)
	Scheme string `json:"scheme"`

	// Location to nginx.conf template - default: /etc/nginx/nginx.template
	Template string `json:"template"`

	// Location of the nginx.conf - default: /etc/nginx/nginx.conf
	NginxConfig string `json:"nginx_config"`

	// User defined configuration data that can be used as part of the template parsing
	// if Beethoven is launched with --root-apps=false .
	Data map[string]interface{}

	/* Internal */
	Version string         `json:"-"`
	context *reloadContext `json:"-"`
}

type SwarmConfig struct {
	// Target connection string for Swarm
	Endpoint string `json:"endpoint"`

	// Network is the name of the network Beethoven should proxy internal requests to.  This is only used
	// if RouteToNode is set to false (the default)
	Network string `json:"network"`

	// RouteToNode will instruct beethoven to route requests to the public address of the Swarm node.  This
	// can be used in scenarios where Beethoven is running outside of the Swarm cluster
	RouteToNode bool

	// Interval to watch for Swarm topology changes
	WatchIntervalSecs int `json:"watch_interval_secs"`

	// TLS Certificate file
	TLSCert string `json:"tls_cert"`
	// TLS Certificate key
	TLSKey string `json:"tls_key"`
	// TLS CA Certificate
	TLSCACert string `json:"tlsca_cert"`
	// Verify TLS
	TLSVerify bool `json:"tls_verify"`
}

type MarathonConfig struct {
	// The URL to Marathon: ex. http://host:8080
	// Enivronment variable: BT_MARATHON_URLS
	Endpoints []string `json:"endpoints" envconfig:"marathon_urls"`

	// The Marathon ID for Beethoven (optional).  If set,
	// will allow for reloading new configuration changes (if using user Data below).
	ServiceId string `json:"service_id"`

	// The basic auth username - if applicable
	// Enivronment variable: BT_USERNAME
	Username string `json:"username" envconfig:"username"`

	// The basic auth password - if applicable
	// Enivronment variable: BT_PASSWORD
	Password string `json:"password" envconfig:"password"`
}

type reloadContext struct {
	server   string
	name     string
	label    string
	profile  string
	filename string
}

var (
	FileNotFound = errors.New("Cannot find the specified config file")
	dryRun       = false
	rootedApps   = true
)

// AddFlags is a hook to add additional CLI Flags
func AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "", "Path and filename of local configuration file. ex: config.yml")
	cmd.Flags().BoolP("remote", "r", false, "Use remote configuraion server")
	cmd.Flags().StringP("server", "s", "", "Remote: URI to remote config server. ex: http://server:8888, env: CONFIG_SERVER")
	cmd.Flags().String("name", "beethoven", "Remote: The name of the app, env: CONFIG_NAME")
	cmd.Flags().String("label", "master", "Remote: The branch to fetch the config from, env: CONFIG_LABEL")
	cmd.Flags().String("profile", "default", "Remote: The profile to use, env: CONFIG_PROFILE")
	cmd.Flags().Bool("dryrun", false, "Bypass NGINX validation/reload -- used for debugging logs")
	cmd.Flags().Bool("root-apps", true, "True by defaults, template context is all apps from marathon.  False, apps is a field in the template as well as config")
}

func LoadConfigFromCommand(cmd *cobra.Command) (*Config, error) {
	remote, _ := cmd.Flags().GetBool("remote")
	config, _ := cmd.Flags().GetString("config")
	dryRun, _ = cmd.Flags().GetBool("dryrun")
	rootedApps, _ = cmd.Flags().GetBool("root-apps")

	if remote {
		server := os.Getenv("CONFIG_SERVER")
		name := os.Getenv("CONFIG_NAME")
		label := os.Getenv("CONFIG_LABEL")
		profile := os.Getenv("CONFIG_PROFILE")

		if server == "" {
			server, _ = cmd.Flags().GetString("server")
		}

		if name == "" {
			name, _ = cmd.Flags().GetString("name")
		}
		if label == "" {
			label, _ = cmd.Flags().GetString("label")
		}
		if profile == "" {
			profile, _ = cmd.Flags().GetString("profile")
		}
		return loadFromRemote(server, name, label, profile)
	}

	if config != "" {
		return loadFromFile(config)
	}

	cfg := new(Config)
	if err := envconfig.Process("bt", cfg); err != nil {
		return nil, fmt.Errorf(EnvErrorFmt, err.Error())
	}

	if len(cfg.MarthonUrls) == 0 {
		return nil, fmt.Errorf(EnvErrorFmt, "BT_MARATHON_URLS not defined")
	}
	return cfg.loadDefaults(), nil
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
	cfg.context = &reloadContext{filename: configFile}
	return cfg.loadDefaults(), nil
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
	cfg.context = &reloadContext{
		server:  server,
		name:    appName,
		label:   label,
		profile: profile,
	}
	return cfg.loadDefaults(), nil
}

/* Config receivers */

// Reload will re-fetch/load the a subset of the configuration and apply it.
// The data applied is "Data" and "FilterRegExStr" values
func (c *Config) Reload() bool {
	newCfg, err := loadConfigFromContext(c.context)
	if err != nil {
		log.Errorf("Error reloading configuration: %s", err.Error())
		return false
	}

	c.Data = newCfg.Data

	if c.FilterRegExStr != newCfg.FilterRegExStr {
		c.FilterRegExStr = newCfg.FilterRegExStr
		c.ParseRegEx()
	}
	log.Info("Configuration successfully reloaded")
	return true
}

func loadConfigFromContext(c *reloadContext) (*Config, error) {
	if c.filename != "" {
		return loadFromFile(c.filename)
	}
	return loadFromRemote(c.server, c.name, c.label, c.profile)
}

// HttpPort is the port we serve the API with
// default 7777 if config port is undefined
func (c *Config) HttpPort() int {
	if c.Port == 0 {
		return 7777
	}
	return c.Port
}

func (c *Config) loadDefaults() *Config {
	if c.NginxConfig == "" {
		c.NginxConfig = DefaultNginxConfPath
	}
	if c.Template == "" {
		c.Template = DefaultNginxTemplatePath
	}
	if c.Scheme == "" {
		c.Scheme = "http"
	}

	if c.Marathon == nil && (c.MarthonUrls != nil && len(c.MarthonUrls) > 0) {
		c.Marathon = &MarathonConfig{
			Endpoints: c.MarthonUrls,
			Username:  c.Username,
			Password:  c.Password,
		}
		if c.SchedulerType == 0 {
			c.SchedulerType = MarathonScheduler
		}
	} else if c.Marathon != nil && c.SchedulerType == 0 {
		c.SchedulerType = MarathonScheduler
	}

	if c.Swarm != nil {
		if c.Swarm.Endpoint == "" {
			c.Swarm.Endpoint = "unix:///var/run/docker.sock"
		}
		if c.SchedulerType == 0 {
			c.SchedulerType = SwarmScheduler
		}
	}

	c.ParseRegEx()
	return c
}

// ParseRegEx validates and parses that the regex is valid. If the FilterRegExpStr is invalid
// the value is emptied and an Error is logged
func (c *Config) ParseRegEx() {
	if c.FilterRegExStr != "" {
		if rx, err := regexp.Compile(c.FilterRegExStr); err != nil {
			log.Error("Error: ignoring user regex filter: %s", err.Error())
			c.FilterRegExStr = ""
		} else {
			c.filterRegEx = rx
		}
	}
}

func (c *Config) IsFilterDefined() bool {
	return c.filterRegEx != nil
}

func (c *Config) Filter() *regexp.Regexp {
	return c.filterRegEx
}

func (c *Config) DryRun() bool {
	return dryRun
}

// IsTemplatedAppRooted means that application from marathon are the actual object during
// template parsing.  If false then applications are a sub-element.
func (c *Config) IsTemplatedAppRooted() bool {
	return rootedApps
}
