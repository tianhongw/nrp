package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/judwhite/go-svc"
	"github.com/spf13/cobra"
	"github.com/tianhongw/grp/client"
	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/util"
	"github.com/tianhongw/grp/version"
)

const (
	defaultCfgFile = "$HOME/.nrp.toml"
	defaultCfgType = "toml"
)

var (
	cfgFile string
	cfgType string
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "nrpc",
		Version: fmt.Sprintf(
			`%s
Git branch: %s
Git commit: %s
Git summary: %s
Commit time: %s
Build time: %s`,
			version.Version,
			version.GitBranch,
			version.GitCommit,
			version.GitSummary,
			version.GitCommitTime,
			version.BuildTime,
		),
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return util.InitProfiling()
		},
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return util.FlushProfiling()
		},
	}

	flags := cmd.PersistentFlags()

	flags.StringVarP(&cfgFile, "config", "c", "", fmt.Sprintf("Config file (default is %s)", defaultCfgFile))
	flags.StringVarP(&cfgType, "type", "t", "", fmt.Sprintf("Config file type (default is %s)", defaultCfgType))

	util.AddProfilingFlags(flags)

	return cmd
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("get home dir failed, ", err)
		}

		cfgFile = strings.Replace(defaultCfgFile, "$HOME", home, 1)
	}

	if cfgType == "" {
		cfgType = defaultCfgType
	}

	if cfg, err := conf.Init(cfgFile, cfgType); err != nil {
		log.Fatal("init config file failed, ", err)
	} else {
		cfgFile = cfg
	}
}

type program struct {
	once sync.Once
	nrpc *client.Client
}

func main() {
	if err := NewCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func serve() {
	prog := &program{}

	if err := svc.Run(prog, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	cfg := conf.GetConfig()

	p.nrpc = client.NewClient(cfg)

	return nil
}

func (p *program) Start() error {
	if err := p.nrpc.Run(); err != nil {
		log.Println(err)
		p.Stop()
		os.Exit(1)
	}

	return nil
}

func (p *program) Stop() error {
	p.once.Do(func() {
		p.nrpc.Exit()
	})
	return nil
}
