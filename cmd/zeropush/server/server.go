package server

import (
	"berty.tech/zero-push/proto"
	"berty.tech/zero-push/server"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"net"
	"os"
)

type serverOptions struct {
	grpcBind         string   `mapstructure:"grpc-bind"`
	apnsCerts        []string `mapstructure:"apns-certs"`
	apnsDevVoipCerts []string `mapstructure:"apns-dev-voip-certs"`
	fcmAPIKeys       []string `mapstructure:"fcm-api-keys"`
	privateKeyFile   string   `mapstructure:"private-key-file"`
	pushJSONKey      string   `mapstructure:"push-json-key"`
}

var currentServerOptions = &serverOptions{}

var Command = &cobra.Command{
	Use:   "server",
	Short: "Starts a push server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runServer(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	cobra.OnInitialize(defaultServerOptions)
	Command.PersistentFlags().StringVar(&currentServerOptions.privateKeyFile, "private-key-file", "", "set private key file for node")
	Command.PersistentFlags().StringVar(&currentServerOptions.grpcBind, "grpc-bind", ":1337", "gRPC listening address")
	Command.PersistentFlags().StringSliceVar(&currentServerOptions.apnsCerts, "apns-certs", []string{}, "Path of APNs certificates, delimited by commas")
	Command.PersistentFlags().StringSliceVar(&currentServerOptions.apnsDevVoipCerts, "apns-dev-voip-certs", []string{}, "Path of APNs VoIP development certificates, delimited by commas")
	Command.PersistentFlags().StringSliceVar(&currentServerOptions.fcmAPIKeys, "fcm-api-keys", []string{}, "API keys for Firebase Cloud Messaging, in the form packageid:token, delimited by commas")
	Command.PersistentFlags().StringVar(&currentServerOptions.pushJSONKey, "push-json-key", "", "In which JSON key the payload should be put")
}

func defaultServerOptions() {
}


func runServer() error {
	lis, err := net.Listen("tcp", currentServerOptions.grpcBind)
	if err != nil {
		return err
	}

	s, err := server.BuildServer(&server.Config{
		GrpcBind:         currentServerOptions.grpcBind,
		ApnsCerts:        currentServerOptions.apnsCerts,
		ApnsDevVoipCerts: currentServerOptions.apnsDevVoipCerts,
		FcmAPIKeys:       currentServerOptions.fcmAPIKeys,
		PrivateKeyFile:   currentServerOptions.privateKeyFile,
		PushJSONKey:      currentServerOptions.pushJSONKey,
	})
	if err != nil {
		return err
	}

	logger().Info(fmt.Sprintf("Starting push server, listening on %s", currentServerOptions.grpcBind))

	grpcServer := grpc.NewServer()
	proto.RegisterPushServiceServer(grpcServer, s)

	grpcServer.Serve(lis)

	return nil
}
