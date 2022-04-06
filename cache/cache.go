package cache

import "time"

type Cache interface {
	Add(key []byte, value interface{}) error
	AddWithExpiration(key []byte, value interface{}, expiration time.Duration) error
	Get(key []byte) (value interface{}, found bool)
	Set(key []byte, value interface{}) error
	SetWithExpiration(key []byte, value interface{}, expiration time.Duration) error
}
