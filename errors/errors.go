package errors

import "errors"

var ErrPushInvalidServerConfig = errors.New("invalid push server config")
var ErrPushMissingBundleId = errors.New("missing bundle id for push")
var ErrPushUnknownDestination = errors.New("invalid push destination")
var ErrPushProvider = errors.New("an error occurred while sending push")
var ErrPushUnknownProvider = errors.New("unknown push type")
var ErrNoProvidersConfigured = errors.New("invalid configuration, no push provider configured")
var ErrDeserialization = errors.New("deserialization error")
var ErrInvalidPrivateKey = errors.New("invalid private key")
