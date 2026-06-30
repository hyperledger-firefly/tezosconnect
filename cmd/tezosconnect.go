package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-common/pkg/i18n"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger-firefly/tezosconnect/internal/tezos"
	fftmcmd "github.com/hyperledger/firefly-transaction-manager/cmd"
	"github.com/hyperledger/firefly-transaction-manager/pkg/fftm"
	txhandlerfactory "github.com/hyperledger/firefly-transaction-manager/pkg/txhandler/registry"
	"github.com/hyperledger/firefly-transaction-manager/pkg/txhandler/simple"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sigs = make(chan os.Signal, 1)

var rootCmd = &cobra.Command{
	Use:   "tezosconnect",
	Short: "Hyperledger FireFly Connector for Tezos blockchain",
	Long:  ``,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx, cancelCtx := context.WithCancel(context.Background())
		return run(ctx, cancelCtx)
	},
}

var cfgFile string

var connectorConfig config.Section

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "f", "", "config file")
	rootCmd.AddCommand(versionCommand())
	rootCmd.AddCommand(configCommand())
	rootCmd.AddCommand(fftmcmd.ClientCommand())
	migrateCmd := fftmcmd.MigrateCommand(func() error {
		InitConfig()
		err := config.ReadConfig("tezosconnect", cfgFile)
		config.SetupLogging(context.Background())
		return err
	})
	migrateCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "config file")
	rootCmd.AddCommand(migrateCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func InitConfig() {
	fftm.InitConfig()
	connectorConfig = config.RootSection("connector")
	tezos.InitConfig(connectorConfig)
	txhandlerfactory.RegisterHandler(&simple.TransactionHandlerFactory{})
}

func run(ctx context.Context, cancelCtx context.CancelFunc) error {
	err := config.ReadConfig("tezosconnect", cfgFile)

	// Setup logging after reading config (even if failed), to output header correctly
	defer cancelCtx()
	ctx = log.WithLogger(ctx, logrus.WithField("pid", fmt.Sprintf("%d", os.Getpid())))
	ctx = log.WithLogger(ctx, logrus.WithField("prefix", "tezosconnect"))

	config.SetupLogging(ctx)

	// Deferred error return from reading config
	if err != nil {
		cancelCtx()
		return i18n.WrapError(ctx, err, i18n.MsgConfigFailed)
	}

	// Init connector
	c, err := tezos.NewTezosConnector(ctx, connectorConfig)
	if err != nil {
		return err
	}
	m, err := fftm.NewManager(ctx, c)
	if err != nil {
		return err
	}

	// Setup signal handling to cancel the context, which shuts down the API Server
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.L(ctx).Infof("Shutting down due to %s", sig.String())
		cancelCtx()
	}()

	return runManager(ctx, m)
}

func runManager(ctx context.Context, m fftm.Manager) error {
	err := m.Start()
	if err != nil {
		return err
	}
	<-ctx.Done()
	m.Close()
	return nil
}
