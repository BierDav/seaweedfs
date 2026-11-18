package command

import (
	"context"
	"errors"
	"github.com/seaweedfs/seaweedfs/weed/pb/filer_pb"
	"google.golang.org/grpc"
	"testing"
)

type backupLiveTestClient struct {
	filer_pb.SeaweedFilerClient
	live *filer_pb.Entry
	err  error
}

func (c *backupLiveTestClient) LookupDirectoryEntry(context.Context, *filer_pb.LookupDirectoryEntryRequest, ...grpc.CallOption) (*filer_pb.LookupDirectoryEntryResponse, error) {
	return &filer_pb.LookupDirectoryEntryResponse{Entry: c.live}, c.err
}
func (c *backupLiveTestClient) WithFilerClient(_ bool, fn func(filer_pb.SeaweedFilerClient) error) error {
	return fn(c)
}
func (*backupLiveTestClient) AdjustedUrl(*filer_pb.Location) string { return "" }
func (*backupLiveTestClient) GetDataCenter() string                 { return "" }
func TestBackupSupersededUpdates(t *testing.T) {
	entry := func(name string, ns int32) *filer_pb.Entry {
		return &filer_pb.Entry{Name: name, Attributes: &filer_pb.FuseAttributes{Mtime: 100, MtimeNs: ns}}
	}
	for _, tc := range []struct {
		name            string
		old, next, live *filer_pb.Entry
		lookupErr       error
		copies          int
		wantErr         bool
	}{
		{name: "skip intermediate update", old: entry("video", 1), next: entry("video", 2), live: entry("video", 3)},
		{name: "copy latest update", old: entry("video", 1), next: entry("video", 3), live: entry("video", 3), copies: 1},
		{name: "copy when source is older", old: entry("video", 1), next: entry("video", 3), live: entry("video", 2), copies: 1},
		{name: "preserve create", next: entry("video", 1), live: entry("video", 3), copies: 1},
		{name: "preserve rename", old: entry("old", 1), next: entry("video", 2), live: entry("video", 3), copies: 1},
		{name: "preserve delete", old: entry("video", 1), copies: 1},
		{name: "lookup failure cannot advance", old: entry("video", 1), next: entry("video", 2), lookupErr: errors.New("source unavailable"), wantErr: true},
		{name: "missing source still applies update", old: entry("video", 1), next: entry("video", 2), copies: 1},
		{name: "unknown timestamp still applies update", old: entry("video", 1), next: &filer_pb.Entry{Name: "video"}, live: entry("video", 3), copies: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copied := 0
			fn := backupLiveFileUpdates(&backupLiveTestClient{live: tc.live, err: tc.lookupErr}, func(*filer_pb.SubscribeMetadataResponse) error { copied++; return nil })
			err := fn(&filer_pb.SubscribeMetadataResponse{Directory: "/uploads", EventNotification: &filer_pb.EventNotification{OldEntry: tc.old, NewEntry: tc.next, NewParentPath: "/uploads"}})
			if (err != nil) != tc.wantErr || copied != tc.copies {
				t.Fatalf("copies=%d err=%v; want copies=%d error=%v", copied, err, tc.copies, tc.wantErr)
			}
		})
	}
}
