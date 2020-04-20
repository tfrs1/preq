package clientutils

import (
	"preq/pkg/bitbucket"
	"preq/pkg/client"
)

type ClientFactory struct{}

func (cf ClientFactory) DefaultClient() (client.Client, error) {
	return bitbucket.DefaultClient()
}
