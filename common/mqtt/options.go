package mqtt

import (
	"flag"
	"github.com/spf13/cobra"
	"time"
)

type Options struct {
	Addr          string        `yaml:"addr,omitempty" json:"addr,omitempty" mapstructure:"addr,omitempty"`
	ClientId      string        `yaml:"cid,omitempty" json:"cid,omitempty" mapstructure:"clientId,omitempty"`
	Un            string        `yaml:"un,omitempty" json:"un,omitempty" mapstructure:"un,omitempty"`
	Pw            string        `yaml:"pw,omitempty" json:"pw,omitempty" mapstructure:"pw,omitempty"`
	Retry         bool          `yaml:"retry,omitempty" json:"retry,omitempty" mapstructure:"retry,omitempty"`
	RetryInterval time.Duration `yaml:"retryInterval,omitempty" json:"retryInterval,omitempty" mapstructure:"retryInterval,omitempty"`
	StorePath     string        `yaml:"fileStorePath,omitempty" json:"fileStorePath,omitempty" mapstructure:"storePath,omitempty"`
	Timeout       time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" mapstructure:"timeout,omitempty"`
	Format        string        `yaml:"format,omitempty" json:"format,omitempty" mapstructure:"format,omitempty"`
	Debug         int           `yaml:"debug,omitempty" json:"debug,omitempty" mapstructure:"debug,omitempty"`
	BasePath      string
}

func (opts *Options) IsEnabled() bool {
	return len(opts.Addr) > 0
}
func (opts *Options) SetDebug(debug int) *Options {
	opts.Debug = debug
	return opts
}
func (opts *Options) SetBasePath(basePath string) *Options {
	opts.BasePath = basePath
	return opts
}

const (
	Format_Raw  = 0
	Format_Json = 1
	Format_Bson = 2
)

func (opts *Options) Parse(bParse bool) {
	flag.StringVar(&opts.Addr, "mqtt.addr", "", "mqtt broker addr")
	flag.StringVar(&opts.ClientId, "mqtt.cid", "", "mqtt client id")
	flag.StringVar(&opts.Un, "mqtt.un", "test", "mqtt client un")
	flag.StringVar(&opts.Pw, "mqtt.pw", "", "mqtt client pw")
	flag.StringVar(&opts.StorePath, "mqtt.store", "", "use mqtt file store")
	flag.BoolVar(&opts.Retry, "mqtt.retry", true, "if retry connect")
	flag.DurationVar(&opts.RetryInterval, "mqtt.retry.duration", 10*time.Second, "retry interval")
	flag.DurationVar(&opts.Timeout, "mqtt.timeout", 30*time.Second, "default timeout")
	flag.StringVar(&opts.Format, "mqtt.format", "bson", "if use public json data")
	if bParse {
		flag.Parse()
	}
}

func (opts *Options) Prepare(c *cobra.Command) {
	c.Flags().StringVar(&opts.Addr, "mqtt.addr", "", "mqtt broker addr")
	c.Flags().StringVar(&opts.ClientId, "mqtt.cid", "", "mqtt client id")
	c.Flags().StringVar(&opts.Un, "mqtt.un", "test", "mqtt client un")
	c.Flags().StringVar(&opts.Pw, "mqtt.pw", "", "mqtt client pw")
	c.Flags().StringVar(&opts.StorePath, "mqtt.store", "", "use mqtt store dir. @mem means using mem store.")
	c.Flags().BoolVar(&opts.Retry, "mqtt.retry", true, "if retry connect")
	c.Flags().DurationVar(&opts.RetryInterval, "mqtt.retry.duration", 10*time.Second, "retry interval")
	c.Flags().DurationVar(&opts.Timeout, "mqtt.timeout", 10*time.Second, "default timeout")
	c.Flags().StringVar(&opts.Format, "mqtt.format", "bson", "if use public json data")
}
