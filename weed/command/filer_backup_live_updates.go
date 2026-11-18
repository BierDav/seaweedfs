package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/seaweedfs/seaweedfs/weed/pb"
	"github.com/seaweedfs/seaweedfs/weed/pb/filer_pb"
	"github.com/seaweedfs/seaweedfs/weed/replication/sink/filersink"
	"github.com/seaweedfs/seaweedfs/weed/util"
)

// A local mirror needs the latest state. Intermediate in-place updates can
// otherwise repeatedly copy a growing upload in full, blocking the stream.
// Creates, deletes and renames must always run to preserve namespace changes.
func backupLiveFileUpdates(client filer_pb.FilerClient, next pb.ProcessMetadataFunc) pb.ProcessMetadataFunc {
	return func(resp *filer_pb.SubscribeMetadataResponse) error {
		if resp == nil || resp.EventNotification == nil || !filer_pb.IsUpdate(resp) || resp.EventNotification.NewEntry.IsDirectory {
			return next(resp)
		}
		incoming := filersink.EntryMtimeNs(resp.EventNotification.NewEntry)
		if incoming <= 0 {
			return next(resp)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		live, _, _, err := filer_pb.GetEntry(ctx, client, util.FullPath(filer_pb.MetadataEventTargetFullPath(resp)))
		if errors.Is(err, filer_pb.ErrNotFound) {
			return next(resp)
		}
		if err != nil {
			return fmt.Errorf("verify current backup source: %w", err)
		}
		// A missing entry alone does not prove that the namespace side effects
		// are safe to omit. Leave those to normal replay and 404 handling.
		if live != nil && !live.IsDirectory && filersink.EntryMtimeNs(live) > incoming {
			return nil
		}
		return next(resp)
	}
}
