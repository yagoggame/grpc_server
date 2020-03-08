/*
Copyright © 2020 Blinnikov AA <goofinator@mail.ru>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/yagoggame/api"
	"github.com/yagoggame/gomaster"
	"github.com/yagoggame/grpc_server/authorization/dummy"
	"github.com/yagoggame/grpc_server/cmd/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "grpc_server",
	Short: "grpc_server is a grpc server of yagogame",
	Long: `grpc_server is a part of yagogame. 
Yagogame - is yet another Go game on the Go, made just for fun.
grpc_server provides go game service available thru grpc`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: runService,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.grpc_server.yaml)")

	rootCmd.PersistentFlags().StringP("address", "a", "localhost", "ip address of grpc_server")
	viper.BindPFlag("address", rootCmd.Flag("address"))
	rootCmd.PersistentFlags().IntP("port", "p", 7777, "port of grpc_server")
	viper.BindPFlag("port", rootCmd.Flag("port"))
	rootCmd.PersistentFlags().StringP("cert", "C", "", "file with TLS certificate")
	viper.BindPFlag("cert", rootCmd.Flag("cert"))
	rootCmd.PersistentFlags().StringP("key", "K", "", "file with TLS key")
	viper.BindPFlag("key", rootCmd.Flag("key"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".grpc_server" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".grpc_server")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func iniFromViper(initData *server.IniDataContainer, command *cobra.Command) {
	initData.Port = viper.GetInt("port")
	initData.IP = viper.GetString("address")
	initData.CertFile = viper.GetString("cert")
	initData.KeyFile = viper.GetString("key")
}

func createServer(initData *server.IniDataContainer) (net.Listener, *grpc.Server) {
	creds, err := credentials.NewServerTLSFromFile(initData.CertFile, initData.KeyFile)
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", initData.IP, initData.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{grpc.Creds(creds),
		grpc.UnaryInterceptor(server.UnaryInterceptor)}

	return lis, grpc.NewServer(opts...)
}

func runService(cmd *cobra.Command, args []string) {
	initData := new(server.IniDataContainer)
	iniFromViper(initData, cmd)

	lis, grpcServer := createServer(initData)

	gamePool := gomaster.NewGamersPool()
	// gameGeter is separated from the object for testing purposes
	gameGeter := server.NewGameGeter(gamePool)
	s := server.NewServer(dummy.New(), gamePool, gameGeter)
	defer s.Release()

	api.RegisterGoGameServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
