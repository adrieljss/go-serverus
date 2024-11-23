package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/adrieljss/go-serverus/env"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// this file contains pgxscan wrappers (especially for caching)

var FetchOne = pgxscan.Get
var FetchMany = pgxscan.Select

// A caching wrapper to
//
//	pgxscan.Get(ctx context.Context, db pgxscan.Querier, dst interface{}, query string, args ...interface{})
//
// Queries to PostgreSQL, but contacts the local Redis DB first.
// Only use in frequent operations (such as comments, posts, etc), as a cache-miss is expensive.
func FetchOneWithCache(ctx context.Context, db pgxscan.Querier, dst interface{}, query string, args ...interface{}) error {
	return fetchWithCache(FetchOne, ctx, db, dst, query, args...)
}

// A caching wrapper to
//
//	pgxscan.Select(ctx context.Context, db pgxscan.Querier, dst interface{}, query string, args ...interface{})
//
// Queries to PostgreSQL, but contacts the local Redis DB first.
// Only use in frequent operations (such as comments, posts, etc), as a cache-miss is expensive.
func FetchManyWithCache(ctx context.Context, db pgxscan.Querier, dst interface{}, query string, args ...interface{}) error {
	return fetchWithCache(FetchMany, ctx, db, dst, query, args...)
}

type pgxscanFunc = func(context.Context, pgxscan.Querier, interface{}, string, ...interface{}) error

func fetchWithCache(fetchFunc pgxscanFunc, ctx context.Context, db pgxscan.Querier, dst interface{}, query string, args ...interface{}) error {
	if env.CEnableRedisCaching {
		h := sha256.New()
		h.Write([]byte(query))
		hash := hex.EncodeToString(h.Sum(nil))
		res, err := RDB.Get(ctx, hash).Result()
		if err == redis.Nil {
			// noop
		} else if err != nil {
			logrus.Errorf("error on redis GET, maybe something is wrong: %s", err)
		} else {
			err := json.Unmarshal([]byte(res), dst)
			if err != nil {
				logrus.Errorf("cannot unmarshal redis result: %s", err)
			} else {
				return nil // if nothing errors out
			}
		}

		err = fetchFunc(ctx, db, dst, query, args...)

		if err != nil {
			return err // return pgError
		}

		// if something errors out or no cache is found.
		// Store a cache
		jsonBytes, err := json.Marshal(dst)
		if err != nil {
			logrus.Errorf("cannot store redis cache: %s", err)
		}
		res, err = RDB.Set(ctx, hash, string(jsonBytes), time.Duration(env.CRedisCacheDuration)*time.Second).Result()
		if err != nil {
			logrus.Errorf("cannot set redis cache: %s", res)
		}
		return nil
	} else {
		return fetchFunc(ctx, db, dst, query, args...)
	}
}
