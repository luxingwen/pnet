package config

import (
	"time"

	"github.com/luxingwen/pnet/protos"
	"github.com/luxingwen/pnet/transport"

	"github.com/imdario/mergo"
)

type Config struct {
	Name                string
	Transport           string
	SupportedTransports []transport.Transport
	Hostname            string
	Port                uint16

	Multiplexer        string
	NumStreamsToOpen   uint32
	NumStreamsToAccept uint32

	LocalHandleMsgWorkers          uint32
	LocalRxMsgCacheExpiration      time.Duration
	LocalRxMsgCacheCleanupInterval time.Duration
	LocalRxMsgChanLen              uint32
	LocalHandleMsgChanLen          uint32

	LocalRxMsgCacheRoutingType     []protos.RoutingType
	LocalReciveMsgCacheRoutingType []protos.RoutingType

	MaxMessageSize           uint32
	DefaultReplyTimeout      time.Duration
	ReplyChanCleanupInterval time.Duration

	RemoteRxMsgChanLen uint32
	RemoteTxMsgChanLen uint32

	RemoteTxMsgCacheExpiration      time.Duration
	RemoteTxMsgCacheCleanupInterval time.Duration
	RemoteTxMsgCacheRoutingType     []protos.RoutingType
	MeasureRoundTripTimeInterval    time.Duration

	KeepAliveTimeout time.Duration
	DialTimeout      time.Duration

	OverlayLocalMsgChanLen uint32
}

func DefaultConfig() *Config {
	return &Config{
		Transport:                      "tcp",
		SupportedTransports:            []transport.Transport{transport.NewTCPTransport()},
		LocalReciveMsgCacheRoutingType: []protos.RoutingType{protos.RELAY},
		Multiplexer:                    "smux",
		NumStreamsToOpen:               8,
		NumStreamsToAccept:             32,

		LocalRxMsgChanLen:     23333,
		LocalHandleMsgWorkers: 1,

		LocalRxMsgCacheExpiration:      300 * time.Second,
		LocalRxMsgCacheCleanupInterval: 10 * time.Second,
		OverlayLocalMsgChanLen:         2333,
		LocalHandleMsgChanLen:          2333,

		RemoteRxMsgChanLen: 2333,
		RemoteTxMsgChanLen: 2333,

		RemoteTxMsgCacheExpiration:      300 * time.Second,
		RemoteTxMsgCacheCleanupInterval: 10 * time.Second,

		MaxMessageSize:               20 * 1024 * 1024,
		DefaultReplyTimeout:          10 * time.Second,
		ReplyChanCleanupInterval:     1 * time.Second,
		MeasureRoundTripTimeInterval: 5 * time.Second,
		KeepAliveTimeout:             50 * time.Second,
		DialTimeout:                  5 * time.Second,
	}
}

func MergedConfig(conf *Config) (*Config, error) {
	merged := DefaultConfig()
	err := mergo.Merge(merged, conf, mergo.WithOverride)
	if err != nil {
		return nil, err
	}
	return merged, nil
}
