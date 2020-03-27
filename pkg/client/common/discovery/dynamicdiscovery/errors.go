package dynamicdiscovery

import (
	"github.com/pkg/errors"
	"strings"
)

const (
	// AccessDenied indicates that the user does not have permission to perform the operation
	AccessDenied = "access denied"
)

// DiscoveryError is an error originating at the Discovery service
type DiscoveryError struct {
	error
	target string
}

func (e DiscoveryError) Error() string {
	return errors.Wrapf(e.error, "target [%s]", e.target).Error()
}

func (e DiscoveryError) Target() string {
	return e.target
}

func (e DiscoveryError) IsAccessDenied() bool {
	return strings.Contains(e.Error(), AccessDenied)
}

func newDiscoveryError(err error, target string) error {
	return DiscoveryError{target: target, error: err}
}
