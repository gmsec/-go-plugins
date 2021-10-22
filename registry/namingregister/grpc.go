// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package namingregister

import (
	"context"
	"fmt"
	"time"

	"github.com/gmsec/micro/naming"
	"github.com/xxjwxc/public/mylog"
	"github.com/xxjwxc/public/tools"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrWatcherClosed = fmt.Errorf("naming: watch closed")

// GRPCResolver creates a grpc.Watcher for a target to track its resolution changes.
type GRPCResolver struct {
	// Client is an initialized etcd client.
	Client       NamingClient
	HeartTimeout time.Duration
}

func (gr *GRPCResolver) Update(ctx context.Context, target string, nm naming.Update) (err error) {
	switch nm.Op {
	case naming.Add:
		err = gr.Client.Put(ctx, target, nm)
	case naming.Delete:
		err = gr.Client.Delete(ctx, target, nm)
	default:
		return status.Error(codes.InvalidArgument, "naming: bad naming op")
	}
	return err
}

func (gr *GRPCResolver) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &gRPCWatcher{c: gr.Client, target: target, serviceName: target, ctx: ctx, cancel: cancel, HeartTimeout: gr.HeartTimeout}
	return w, nil
}

type gRPCWatcher struct {
	c      NamingClient
	target string
	ctx    context.Context
	cancel context.CancelFunc
	//wch          etcd.WatchChan
	err          error
	HeartTimeout time.Duration
	serviceName  string
}

// Next gets the next set of updates from the etcd resolver.
// Calls to Next should be serialized; concurrent calls are not safe since
// there is no way to reconcile the update ordering.
func (gw *gRPCWatcher) Next() ([]*naming.Update, error) {
	if !gw.c.Watchering() {
		// first Next() returns all addresses
		return gw.firstNext()
	}
	if gw.err != nil {
		return nil, gw.err
	}

	// process new events on target/*
	list, err := gw.c.WatcherNext()
	if err != nil {
		return nil, gw.err
	}

	offset := float64(time.Now().Unix()) - gw.HeartTimeout.Seconds()

	var deleteUpdate []*naming.Update
	updates := make([]*naming.Update, 0, len(list))
	for _, jupdate := range list {
		me, b := jupdate.Metadata.(float64)
		if b && me > offset {
			updates = append(updates, jupdate)
		} else {
			deleteUpdate = append(deleteUpdate, jupdate)
		}
	}

	if len(deleteUpdate) > 0 {
		mylog.Debugf("delete(%v):%v", gw.serviceName, tools.JSONDecode(deleteUpdate))
		gr := &GRPCResolver{Client: gw.c, HeartTimeout: gw.HeartTimeout}
		for _, v := range deleteUpdate {
			gr.Update(context.TODO(), gw.serviceName, naming.Update{
				Op:   naming.Delete,
				Addr: v.Addr,
				// Metadata: r.port,
			})
		}
	}

	return updates, nil
}

func (gw *gRPCWatcher) firstNext() ([]*naming.Update, error) {
	// Use serialized request so resolution still works if the target etcd
	// server is partitioned away from the quorum.
	resp, err := gw.c.Get(gw.ctx, gw.target)
	if gw.err = err; err != nil {
		return nil, err
	}

	offset := float64(time.Now().Unix()) - gw.HeartTimeout.Seconds()
	var deleteUpdate []*naming.Update
	updates := make([]*naming.Update, 0, len(resp))
	for _, jupdate := range resp {
		me, b := jupdate.Metadata.(float64)
		if b && me > offset {
			updates = append(updates, jupdate)
		} else {
			deleteUpdate = append(deleteUpdate, jupdate)
		}
	}

	if len(deleteUpdate) > 0 {
		gr := &GRPCResolver{Client: gw.c, HeartTimeout: gw.HeartTimeout}
		for _, v := range deleteUpdate {
			mylog.Debugf("delete(%v):%v", gw.serviceName, tools.JSONDecode(deleteUpdate))
			gr.Update(context.TODO(), gw.serviceName, naming.Update{
				Op:   naming.Delete,
				Addr: v.Addr,
				// Metadata: r.port,
			})
		}
	}

	gw.c.Watch(gw.ctx, gw.target)
	return updates, nil
}

func (gw *gRPCWatcher) Close() {
	gw.c.Close()
	gw.cancel()
}
