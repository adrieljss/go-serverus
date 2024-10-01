package utils

import (
	"math"
	"sync"
	"time"
)

type item[T any] struct {
	value T
	exp   int64 // exp date in UNIX time
}

// thread safe map with a ttl
type TtlMap[K comparable, V any] struct {
	mp map[K]*item[V]
	l  sync.Mutex
}

// always use this
// constructs a TtlMap
//
// a go routine is called upon this function call, where
// every `interval` it checks for expired elements
//
// interval to clear cache is best after every 2 hours
//
// ttl is handled by Get() also, if the ttl already expired, Get() will return nil
func NewTtlMap[K comparable, V any](clearInterval time.Duration) (m *TtlMap[K, V]) {
	var x TtlMap[K, V]
	x.mp = make(map[K]*item[V])
	x.__StartObliterator(clearInterval)
	return &x
}

func (mp *TtlMap[K, V]) Len() int {
	return len(mp.mp)
}

// Thread-safe store
//
// exp as the date to expire in UNIX time,
// set exp to -1 to set for infinite time
func (mp *TtlMap[K, V]) Store(key K, val V, exp int64) {
	mp.l.Lock()
	if exp == -1 {
		exp = int64(math.MaxInt64)
	}
	mp.mp[key] = &item[V]{
		value: val,
		exp:   exp,
	}
	mp.l.Unlock()
}

// if ttl already expired (but still exists in cache), this will delete the key and return (nil, false)
func (mp *TtlMap[K, V]) Get(key K) (*V, bool) {
	// no need for mutex lock on read
	if _, ok := mp.mp[key]; !ok {
		return nil, false
	}

	if mp.mp[key].exp < time.Now().Unix() {
		mp.l.Lock()
		delete(mp.mp, key)
		mp.l.Unlock()
		return nil, false
	}
	return &mp.mp[key].value, true
}

func (mp *TtlMap[K, V]) Delete(key K) {
	mp.l.Lock()
	delete(mp.mp, key)
	mp.l.Unlock()
}

// start the interval delete checker for ttlmap,
// this launches a go routine to clear the cache every `interval`
//
// # only call once!
func (mp *TtlMap[K, V]) __StartObliterator(interval time.Duration) {
	go func() {
		for now := range time.Tick(interval) {
			mp.l.Lock()
			for k, v := range mp.mp {
				if v.exp < now.Unix() {
					delete(mp.mp, k)
				}
			}
			mp.l.Unlock()
		}
	}()
}
