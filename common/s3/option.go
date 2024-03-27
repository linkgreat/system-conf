package s3

import (
	"flag"
	"github.com/spf13/cobra"
	"time"
)

type MinioOptions struct {
	Addr       string        `mapstructure:"addr" yaml:"addr,omitempty" json:"addr,omitempty"`
	Access     string        `mapstructure:"access,omitempty" yaml:"access,omitempty" json:"access,omitempty"`
	Secret     string        `mapstructure:"secret,omitempty" yaml:"secret,omitempty" json:"secret,omitempty"`
	Secure     bool          `mapstructure:"secure,omitempty" yaml:"secure,omitempty" json:"secure,omitempty"`
	Region     string        `mapstructure:"region,omitempty" yaml:"region,omitempty" json:"region,omitempty"`
	SkipVerify bool          `mapstructure:"skipVerify,omitempty" yaml:"skipVerify,omitempty" json:"skipVerify,omitempty"`
	Timeout    time.Duration `mapstructure:"timeout,omitempty" yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Debug      int           `mapstructure:"debug,omitempty" yaml:"debug,omitempty" json:"debug,omitempty"`
}

func (opts *MinioOptions) Prepare(c *cobra.Command) {
	c.Flags().StringVar(&opts.Addr, "minio.addr", "", "minio addr")
	c.Flags().StringVar(&opts.Access, "minio.access", "test", "minio un")
	c.Flags().StringVar(&opts.Secret, "minio.secret", "test123456", "minio pw")
	c.Flags().BoolVar(&opts.Secure, "minio.secure", false, "if secure")
	c.Flags().StringVar(&opts.Region, "minio.region", "cn-east-rack0", "region")
	c.Flags().DurationVar(&opts.Timeout, "minio.timeout", 15*time.Second, "minio default timeout")
}
func (opts *MinioOptions) Parse(bParse bool) {
	flag.StringVar(&opts.Addr, "minio.addr", "", "minio addr")
	flag.StringVar(&opts.Access, "minio.access", "test", "minio un")
	flag.StringVar(&opts.Secret, "minio.secret", "test123456", "minio pw")
	flag.BoolVar(&opts.Secure, "minio.secure", false, "if secure")
	flag.DurationVar(&opts.Timeout, "minio.timeout", 15*time.Second, "minio default timeout")
	flag.StringVar(&opts.Region, "minio.region", "", "region")
	if bParse {
		flag.Parse()
	}
}

func (opts *MinioOptions) SetDebug(debug int) *MinioOptions {
	opts.Debug = debug
	return opts
}
