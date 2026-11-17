package localsink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seaweedfs/seaweedfs/weed/pb/filer_pb"
)

// Directory metadata jobs may overlap descendant copies. A missing directory
// must not request genProcessFunction's delete/create fallback: a descendant
// can create it before that delete, then lose it before opening its temp file.
func TestMissingDirectoryUpdateDoesNotRequestDeleteFallback(t *testing.T) {
	dir := t.TempDir()
	sink := &LocalSink{}
	sink.initialize(dir, false)
	key := filepath.Join(dir, "thumbs", "new-directory")
	entry := &filer_pb.Entry{Name: "new-directory", IsDirectory: true}
	handled, err := sink.UpdateEntry(key, entry, filepath.Dir(key), entry, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Interleave a descendant writer after UpdateEntry but before its caller
	// considers the missing-entry fallback, exactly as the processor allows.
	if err := os.MkdirAll(key, 0755); err != nil {
		t.Fatal(err)
	}
	if !handled {
		if err := sink.DeleteEntry(key, true, false, nil); err != nil {
			t.Fatal(err)
		}
	}
	temp := filepath.Join(key, ".seaweedfs-tmp-regression")
	if err := os.WriteFile(temp, []byte("copy"), 0600); err != nil {
		t.Fatalf("directory metadata update removed a descendant writer's directory: %v", err)
	}
	if !handled {
		t.Fatal("directory update requested destructive missing-entry fallback")
	}
}
