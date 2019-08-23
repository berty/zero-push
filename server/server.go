package server

import (
	"context"
	"fmt"
	"strings"

	push "berty.tech/zero-push"
	"berty.tech/zero-push/errors"
	proto_push "berty.tech/zero-push/proto/push"
	proto_service "berty.tech/zero-push/proto/service"
	"berty.tech/zero-push/providers/apns"
	"berty.tech/zero-push/providers/fcm"
)

type Config struct {
	GrpcBind         string   `protobuf:"bytes,1,opt,name=grpcBind,proto3" json:"grpcBind,omitempty"`
	ApnsCerts        []string `protobuf:"bytes,2,rep,name=apnsCerts,proto3" json:"apnsCerts,omitempty"`
	ApnsDevVoipCerts []string `protobuf:"bytes,3,rep,name=apnsDevVoipCerts,proto3" json:"apnsDevVoipCerts,omitempty"`
	FcmAPIKeys       []string `protobuf:"bytes,4,rep,name=fcmAPIKeys,proto3" json:"fcmAPIKeys,omitempty"`
	PrivateKeyFile   string   `protobuf:"bytes,5,opt,name=privateKeyFile,proto3" json:"privateKeyFile,omitempty"`
	PushJSONKey      string   `protobuf:"bytes,6,opt,name=pushJSONKey,proto3" json:"pushJSONKey,omitempty"`
}

func BuildServer(cfg *Config) (*Server, error) {
	dispatchers, err := listPushDispatchers(cfg)
	if err != nil {
		return nil, err
	}

	if len(dispatchers) == 0 {
		return nil, errors.ErrNoProvidersConfigured
	}

	privKey, err := push.LoadAndParsePrivateKey(cfg.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	pushManager := push.NewManager(privKey, dispatchers...)

	return &Server{
		pushManager: pushManager,
		config:      cfg,
	}, nil
}

func listPushDispatchers(cfg *Config) ([]push.Dispatcher, error) {
	var pushDispatchers []push.Dispatcher
	for _, certs := range []struct {
		Certs    []string
		ForceDev bool
	}{
		{Certs: cfg.ApnsCerts, ForceDev: false},
		{Certs: cfg.ApnsDevVoipCerts, ForceDev: true},
	} {
		for _, cert := range certs.Certs {
			dispatcher, err := apns.NewAPNSDispatcher(cert, certs.ForceDev, cfg.PushJSONKey)
			if err != nil {
				return nil, err
			}

			logger().Info(fmt.Sprintf("registering apns provider for path %s", cert))

			pushDispatchers = append(pushDispatchers, dispatcher)
		}
	}

	for _, apiKey := range cfg.FcmAPIKeys {
		dispatcher, err := fcm.NewFCMDispatcher(apiKey, cfg.PushJSONKey)
		if err != nil {
			return nil, err
		}

		appBundle := strings.Split(apiKey, ":")[0]

		logger().Info(fmt.Sprintf("registering fcm provider for app %s", appBundle))

		pushDispatchers = append(pushDispatchers, dispatcher)
	}

	return pushDispatchers, nil
}

type Server struct {
	pushManager *push.Manager
	config      *Config
}

func (s *Server) PushTo(ctx context.Context, pushToInput *proto_push.PushToInput) (*proto_push.Void, error) {
	for _, pushData := range pushToInput.PushData {
		if err := s.pushManager.PushTo(ctx, pushData); err != nil {
			return &proto_push.Void{}, err
		}
	}

	return &proto_push.Void{}, nil
}

var _ proto_service.PushServiceServer = (*Server)(nil)
