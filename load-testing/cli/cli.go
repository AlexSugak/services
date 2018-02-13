package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/skycoin/skycoin/src/api/webrpc"
	gcli "github.com/urfave/cli"
)

const (
	Version           = "0.0.1"
	defaultRpcAddress = "127.0.0.1:6430"
)

type Config struct {
	RpcAddress string
}

type App struct {
	gcli.App
}

func LoadConfig() (Config, error) {
	rpcAddr := os.Getenv("RPC_ADDR")
	if rpcAddr == "" {
		rpcAddr = defaultRpcAddress
	}

	return Config{
		RpcAddress: rpcAddr,
	}, nil
}

// NewApp creates a load testing app instance
func NewApp(c Config) *App {
	gcliApp := gcli.NewApp()
	app := &App{
		App: *gcliApp,
	}

	commands := []gcli.Command{
		randomTxsCmd(),
	}

	app.Name = "load-testing"
	app.Version = Version
	app.Usage = "the Skycoin load testing command line interface"
	app.Commands = commands
	app.CommandNotFound = func(ctx *gcli.Context, command string) {
		tmp := fmt.Sprintf("{{.HelpName}}: '%s' is not a {{.HelpName}} command. See '{{.HelpName}} --help'.\n", command)
		gcli.HelpPrinter(app.Writer, tmp, app)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	app.Metadata = map[string]interface{}{
		"config": c,
		"rpc": &webrpc.Client{
			Addr: c.RpcAddress,
		},
		"sigs": &sigs,
	}

	return app
}

// Run starts the load testing cli app
func (app *App) Run(args []string) error {
	rpc := app.Metadata["rpc"].(*webrpc.Client)
	config := app.Metadata["config"].(Config)

	// check node status
	if s, err := rpc.GetStatus(); err != nil {
		return fmt.Errorf("error connecting to node: %s", err)
	} else if !s.Running {
		return fmt.Errorf("node is not running at %s", config.RpcAddress)
	}

	return app.App.Run(args)
}

// SigsFromContext gets os signals channel from cli app context
func SigsFromContext(c *gcli.Context) *chan os.Signal {
	return c.App.Metadata["sigs"].(*chan os.Signal)
}

// RPCClientFromContext gets webrpc client from cli app context
func RPCClientFromContext(c *gcli.Context) *webrpc.Client {
	return c.App.Metadata["rpc"].(*webrpc.Client)
}

func onCommandUsageError(command string) gcli.OnUsageErrorFunc {
	return func(c *gcli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(c.App.Writer, "Error: %v\n\n", err)
		gcli.ShowCommandHelp(c, command)
		return nil
	}
}
