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
	"github.com/yagoggame/grpc_server/authorization/filemap"
	"github.com/yagoggame/grpc_server/cmd/server"
	"github.com/yagoggame/grpc_server/interfaces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile                  string
	acceptedAuthorizator     = []string{"dummy", "filemap"}
	acceptedAuthorizatorFlag = newOfist(acceptedAuthorizator)
)

type oflist struct {
	value          string
	acceptedValues []string
}

func newOfist(acceptedValues []string) *oflist {
	return &oflist{acceptedValues: acceptedValues}
}

func (av *oflist) Set(val string) error {
	if !isInList(val, av.acceptedValues) {
		return fmt.Errorf("expected flag value of list: %v. got: %v", av.acceptedValues, val)
	}
	av.value = val
	return nil
}

func (av *oflist) Type() string {
	return "oflist"
}

func (av *oflist) String() string { return av.value }

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

	rootCmd.PersistentFlags().VarP(acceptedAuthorizatorFlag, "authorizator", "A", fmt.Sprintf("one of %v values to chose authorizator", acceptedAuthorizator))
	viper.BindPFlag("authorizator", rootCmd.Flag("authorizator"))
	rootCmd.PersistentFlags().StringP("filename", "F", "", "filename to be used by filemap authorizator")
	viper.BindPFlag("filename", rootCmd.Flag("filename"))
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

	initData.Authorizer = viper.GetString("authorizator")
	if err := acceptedAuthorizatorFlag.Set(initData.Authorizer); err != nil {
		log.Fatalf("Error: invalid argument %v for \"-A, --authorizator\" flag:%v\n%s", initData.Authorizer, err, command.UsageString())
	}

	initData.Filename = viper.GetString("filename")
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
	authorizator := getAuthorizator(initData)
	gameGeter := server.NewGameGeter(gamePool)
	s := server.NewServer(authorizator, gamePool, gameGeter)
	defer s.Release()

	api.RegisterGoGameServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func getAuthorizator(initData *server.IniDataContainer) interfaces.Authorizator {
	switch initData.Authorizer {
	case "dummy":
		return dummy.New()
	case "filemap":
		authorizator, err := filemap.New(initData.Filename)
		if err != nil {
			log.Fatalf("failed to create filemap authorizator: %s", err)
		}
		return authorizator
	}
	log.Fatalf("failed to create %q authorizator of unknown type", initData.Authorizer)
	return nil
}

func isInList(str string, list []string) bool {
	for _, variant := range list {
		if str == variant {
			return true
		}
	}
	return false
}
