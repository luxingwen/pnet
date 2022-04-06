package pnet

import (
	"pnet/config"
	"pnet/log"
	"pnet/node"
	"pnet/overlay"
	"pnet/overlay/lnet"

	"github.com/google/uuid"
)

type PNet struct {
	overlay.Network
}

func NewPNet(id string, conf *config.Config) (*PNet, error) {
	var mergedConf *config.Config
	var err error

	if conf != nil {
		convertedConf := config.Config(*conf)
		mergedConf, err = config.MergedConfig(&convertedConf)
		if err != nil {
			return nil, err
		}
	} else {
		mergedConf = config.DefaultConfig()
	}

	if id == "" {
		id = uuid.New().String()
	}

	localNode, err := node.NewLocalNode(id, mergedConf)
	if err != nil {
		return nil, err
	}

	network, err := lnet.NewLNet(localNode)

	pn := &PNet{
		Network: network,
	}

	return pn, nil

}

// SetLogger sets the global logger
func SetLogger(logger log.Logger) error {
	return log.SetLogger(logger)
}
