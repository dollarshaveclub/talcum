package talcum

import "github.com/hashicorp/consul/api"

// ConsulKVClient is the interface to a Consul KV store with a
// check-and-set operation.
type ConsulKVClient interface {
	CAS(p *api.KVPair, q *api.WriteOptions) (bool, *api.WriteMeta, error)
}

// ConsulLocker can lock keys using Consul as a backend.
type ConsulLocker struct {
	kvClient ConsulKVClient
}

// Lock tries to lock a key, return true if the lock operation was
// successful.
func (c *ConsulLocker) Lock(key string) (bool, error) {
	set, _, err := c.kvClient.CAS(&api.KVPair{
		Key:   key,
		Value: []byte("1"),
	}, nil)
	if err != nil {
		return false, err
	}
	return set, nil
}
