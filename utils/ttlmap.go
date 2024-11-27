package utils

import (
	"sync"
	"time"
)

type item[T any] struct {
	value           T
	initialDuration int64
	exp             int64 // exp date in UNIX time
}

// thread safe map with a ttl
type TtlMap[K comparable, V any] struct {
	mp map[K]*item[V]
	l  sync.Mutex
	is bool // is this a fix-date ttl map (0) or last-access ttl map (1)
}

// Constructs a Non-LastAccess TtlMap.
// A go routine is called upon this function call, where
// every `interval` it checks for expired elements.
// Interval to clear cache is best after every 2 hours as to not put too much load on the server.
// TTL is handled by Get() also, if the ttl already expired, Get() will return nil.
func NewTtlMap[K comparable, V any](clearInterval time.Duration) (m *TtlMap[K, V]) {
	var x TtlMap[K, V]
	x.mp = make(map[K]*item[V])
	x.is = false
	x.__StartObliterator(clearInterval)
	return &x
}

// Constructs a LastAccess TtlMap.
// A go routine is called upon this function call, where
// every `interval` it checks for expired elements.
// Interval to clear cache is best after every 2 hours as to not put too much load on the server.
// TTL is handled by Get() also, if the ttl already expired, Get() will return nil.
// LastAccess TTLMap will refresh the TTL when accessed. (will update the exp date to now() + initialDuration)
func NewLastAccessTtlMap[K comparable, V any](clearInterval time.Duration) (m *TtlMap[K, V]) {
	var x TtlMap[K, V]
	x.mp = make(map[K]*item[V])
	x.is = true
	x.__StartObliterator(clearInterval)
	return &x
}

func (mp *TtlMap[K, V]) Len() int {
	return len(mp.mp)
}

// Thread-safe store.
// Exp as the date to expire in UNIX time. Set expOffset to -1 to set for infinite time (may cause huge memory being used).
// The final exp time is now() + expOffset.
func (mp *TtlMap[K, V]) Store(key K, val V, expOffset int64) {
	mp.l.Lock()
	if expOffset == -1 {
		mp.mp[key] = &item[V]{
			value:           val,
			exp:             -1,
			initialDuration: -1,
		}
	} else {
		mp.mp[key] = &item[V]{
			value:           val,
			exp:             time.Now().Unix() + expOffset,
			initialDuration: expOffset,
		}
	}
	mp.l.Unlock()
}

// If ttl already expired (but still exists in cache), this will delete the key and return (nil, false).
// If this is a last access map and the item still exists, this will update exp time to now() + initialDuration
func (mp *TtlMap[K, V]) Get(key K) (*V, bool) {
	// no need for mutex lock on read
	if _, ok := mp.mp[key]; !ok {
		return nil, false
	}

	// handle for -1 (infinite time)
	if mp.mp[key].exp == -1 {
		return &mp.mp[key].value, true
	}

	// if expires, then delete immediately without obliterator.
	if mp.mp[key].exp < time.Now().Unix() {
		mp.l.Lock()
		delete(mp.mp, key)
		mp.l.Unlock()
		return nil, false
	}

	// if is last access map, add to ttl time since last access.
	if mp.is && mp.mp[key].initialDuration >= 0 {
		mp.mp[key].exp = time.Now().Unix() + mp.mp[key].initialDuration
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
				if v.exp >= 0 && v.exp < now.Unix() {
					delete(mp.mp, k)
				}
			}
			mp.l.Unlock()
		}
	}()
}
