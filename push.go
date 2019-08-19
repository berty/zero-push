package push

import (
	zpErrors "berty.tech/zero-push/errors"
	"berty.tech/zero-push/proto"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Dispatcher interface {
	CanDispatch(*proto.PushDestination) bool
	Dispatch(*proto.PushData, *proto.PushDestination) error
}

type Manager struct {
	dispatchers    []Dispatcher
	tokenDecrypter crypto.Decrypter
}

func (m *Manager) Dispatch(pushData *proto.PushData, pushDestination *proto.PushDestination) error {
	var err error
	logger().Info("dispatch push")
	for _, dispatcher := range m.dispatchers {
		if !dispatcher.CanDispatch(pushDestination) {
			continue
		}

		if err = dispatcher.Dispatch(pushData, pushDestination); err == nil {
			return nil
		}
	}

	if err != nil {
		return err
	}

	return zpErrors.ErrPushUnknownProvider
}

func (m *Manager) PushTo(ctx context.Context, pushAttrs *proto.PushData) error {
	logger().Info("Sending push to device")
	identifier, err := m.Decrypt(pushAttrs.PushIdentifier)

	if err != nil {
		return errors.Wrap(err, zpErrors.ErrPushUnknownDestination.Error())
	}

	pushDestination := &proto.PushDestination{}

	if err := pushDestination.Unmarshal(identifier); err != nil {
		return errors.Wrap(err, zpErrors.ErrPushUnknownDestination.Error())
	}

	if err := m.Dispatch(&proto.PushData{
		PushIdentifier: pushAttrs.PushIdentifier,
		Envelope:       pushAttrs.Envelope,
		Priority:       pushAttrs.Priority,
	}, pushDestination); err != nil {
		return err
	}

	return nil
}

func (m *Manager) Decrypt(msg []byte) (plaintext []byte, err error) {
	return m.tokenDecrypter.Decrypt(rand.Reader, msg, nil)
}

func LoadAndParsePrivateKey(path string) (*rsa.PrivateKey, error) {
	privPem, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privPem)

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, zpErrors.ErrInvalidPrivateKey
	}

	return rsaKey, nil
}

func NewManager(tokenDecrypter crypto.Decrypter, dispatchers ...Dispatcher) *Manager {
	pushManager := &Manager{
		dispatchers:    dispatchers,
		tokenDecrypter: tokenDecrypter,
	}

	return pushManager
}
