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

package grpc

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type PoolOptions struct {
	Size *int
	TTL  *time.Duration
}

type pool struct {
	sync.Mutex

	size       int
	ttl        int64
	maxStreams int
	maxIdle    int
	conns      map[string]*streamsPool
}

type streamsPool struct {
	head  *poolConn
	busy  *poolConn
	count int
	idle  int
}

type poolConn struct {
	*grpc.ClientConn

	err  error
	addr string

	pool    *pool
	sp      *streamsPool
	streams int
	created int64

	pre  *poolConn
	next *poolConn
	in   bool
}

func newPool() *pool {
	return &pool{
		size:       1,
		ttl:        int64(0),
		maxStreams: 1,
		maxIdle:    0,
		conns:      make(map[string]*streamsPool),
	}
}

func (p *pool) Init(opts PoolOptions) {
	p.Lock()
	defer p.Unlock()
	if opts.Size != nil {
		p.size = *opts.Size
	}
	if opts.TTL != nil {
		p.ttl = int64(opts.TTL.Seconds())
	}
}

func (p *pool) getConn(ctx context.Context, addr string, opts ...grpc.DialOption) (*poolConn, error) {
	now := time.Now().Unix()
	p.Lock()
	sp, ok := p.conns[addr]
	if !ok {
		sp = &streamsPool{head: &poolConn{}, busy: &poolConn{}, count: 0, idle: 0}
		p.conns[addr] = sp
	}

	conn := sp.head.next

	for conn != nil {
		switch conn.GetState() {
		case connectivity.Connecting:
			conn = conn.next
			continue
		case connectivity.Shutdown:
			next := conn.next
			if conn.streams == 0 {
				removeConn(conn)
				sp.idle--
			}
			conn = next
			continue
		case connectivity.TransientFailure:
			next := conn.next
			if conn.streams == 0 {
				removeConn(conn)
				conn.ClientConn.Close()
				sp.idle--
			}
			conn = next
			continue
		case connectivity.Ready:
		case connectivity.Idle:
		}

		if now-conn.created > p.ttl {
			next := conn.next
			if conn.streams == 0 {
				removeConn(conn)
				conn.ClientConn.Close()
				sp.idle--
			}
			conn = next
			continue
		}

		if conn.streams >= p.maxStreams {
			next := conn.next
			removeConn(conn)
			addConnAfter(conn, sp.busy)
			conn = next
			continue
		}

		if conn.streams == 0 {
			sp.idle--
		}

		conn.streams++
		p.Unlock()
		return conn, nil
	}

	p.Unlock()

	cc, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}
	conn = &poolConn{cc, nil, addr, p, sp, 1, time.Now().Unix(), nil, nil, false}

	p.Lock()
	if sp.count < p.size {
		addConnAfter(conn, sp.head)
	}
	p.Unlock()

	return conn, nil
}

func (p *pool) release(addr string, conn *poolConn, err error) {
	p.Lock()
	p, sp, created := conn.pool, conn.sp, conn.created

	if !conn.in && sp.count < p.size {
		addConnAfter(conn, sp.head)
	}
	if !conn.in {
		p.Unlock()
		conn.ClientConn.Close()
		return
	}

	if conn.streams >= p.maxStreams {
		removeConn(conn)
		addConnAfter(conn, sp.head)
	}
	conn.streams--

	if conn.streams == 0 {
		now := time.Now().Unix()
		if err != nil || sp.idle >= p.maxIdle || now-created > p.ttl {
			removeConn(conn)
			p.Unlock()
			conn.ClientConn.Close()
			return
		}
		sp.idle++
	}
	p.Unlock()
}

func (conn *poolConn) Close() {
	conn.pool.release(conn.addr, conn, conn.err)
}

func removeConn(conn *poolConn) {
	if conn.pre != nil {
		conn.pre.next = conn.next
	}
	if conn.next != nil {
		conn.next.pre = conn.pre
	}
	conn.pre = nil
	conn.next = nil
	conn.in = false
	conn.sp.count--
}

func addConnAfter(conn *poolConn, after *poolConn) {
	conn.next = after.next
	conn.pre = after
	if after.next != nil {
		after.next.pre = conn
	}
	after.next = conn
	conn.in = true
	conn.sp.count++
}
