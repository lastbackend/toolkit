/*
Copyright [2014] - [2022] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/lastbackend/toolkit/pkg/probe/types"
	"time"
)

func PostgresPingChecker(database *sql.DB, timeout time.Duration) types.ProbeFunc {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if database == nil {
			return fmt.Errorf("connection is nil")
		}
		return database.PingContext(ctx)
	}
}

func RedisClientPingChecker(client *redis.Client, timeout time.Duration) types.ProbeFunc {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if client == nil {
			return fmt.Errorf("connection is nil")
		}
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
		return nil
	}
}

func RedisClusterPingChecker(client *redis.ClusterClient, timeout time.Duration) types.ProbeFunc {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if client == nil {
			return fmt.Errorf("connection is nil")
		}
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return err
		}
		return nil
	}
}
