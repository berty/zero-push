package apns

import (
	"encoding/base64"
	"github.com/berty/zero-push"
	zpErrors "github.com/berty/zero-push/errors"
	"github.com/berty/zero-push/proto"
	"strings"

	"github.com/pkg/errors"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

const asn1UID = "0.9.2342.19200300.100.1.1"
const appleCertDevNamePart = "Apple Development IOS Push Services"

type APNSDispatcher struct {
	bundleID    string
	client      *apns2.Client
	jsonDataKey string
}

var _ push.Dispatcher = &APNSDispatcher{}

func NewAPNSDispatcher(path string, forceDev bool, jsonDataKey string) (push.Dispatcher, error) {
	cert, err := certificate.FromP12File(path, "")

	if err != nil {
		return nil, zpErrors.ErrPushInvalidServerConfig
	}

	bundleID := ""

	for _, kv := range cert.Leaf.Subject.Names {
		if kv.Type.String() == asn1UID {
			bundleID = kv.Value.(string)
			break
		}
	}

	if bundleID == "" {
		return nil, zpErrors.ErrPushMissingBundleId
	}

	production := !strings.Contains(cert.Leaf.Subject.CommonName, appleCertDevNamePart)

	client := apns2.NewClient(cert)

	if !forceDev && production {
		client = client.Production()
	} else {
		client = client.Development()
	}

	dispatcher := &APNSDispatcher{
		bundleID:    bundleID,
		client:      client,
		jsonDataKey: jsonDataKey,
	}

	return dispatcher, nil
}

func (d *APNSDispatcher) CanDispatch(pushDestination *proto.PushDestination) bool {
	if pushDestination.PushType != proto.DevicePushType_APNS {
		return false
	}

	apnsIdentifier := &proto.PushNativeIdentifier{}
	if err := apnsIdentifier.Unmarshal(pushDestination.PushId); err != nil {
		return false
	}

	if d.bundleID != apnsIdentifier.PackageID && d.bundleID != apnsIdentifier.PackageID+".voip" {
		return false
	}

	return true
}

func (d *APNSDispatcher) Dispatch(pushData *proto.PushData, pushDestination *proto.PushDestination) error {
	apnsIdentifier := &proto.PushNativeIdentifier{}
	if err := apnsIdentifier.Unmarshal(pushDestination.PushId); err != nil {
		return errors.Wrap(err, zpErrors.ErrPushUnknownDestination.Error())
	}

	pushPayload := payload.NewPayload()
	pushPayload.Custom(d.jsonDataKey, base64.StdEncoding.EncodeToString(pushData.Envelope))
	pushPayload.ContentAvailable()
	notification := &apns2.Notification{}
	notification.DeviceToken = apnsIdentifier.DeviceToken
	notification.Topic = d.bundleID
	notification.Payload = pushPayload

	response, err := d.client.Push(notification)

	if err != nil {
		return errors.Wrap(err, zpErrors.ErrPushProvider.Error())
	} else if response.StatusCode != 200 {
		return errors.Wrap(errors.New(response.Reason), zpErrors.ErrPushProvider.Error())
	}

	return nil
}
