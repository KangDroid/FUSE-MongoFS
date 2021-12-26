package main

import (
	"filesystem/mongonode"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"io/ioutil"
	"log"
)

// inMemoryFS is the root of the tree
type inMemoryFS struct {
	fs.Inode
}

//
//// Ensure that we implement NodeOnAdder
//var _ = (fs.NodeOnAdder)((*inMemoryFS)(nil))
//
//// OnAdd is called on mounting the file system. Use it to populate
//// the file system tree.
//func (root *inMemoryFS) OnAdd(ctx context.Context) {
//	fileList := mongocom.GetAllFileMetadata()
//	log.Printf("Got: %d of metadatas\n", len(fileList))
//
//	for _, eachFile := range fileList {
//		name := eachFile.FileName
//
//		dir, base := filepath.Split(name)
//
//		p := &root.Inode
//
//		// Add directories leading up to the file.
//		for _, component := range strings.Split(dir, "/") {
//			if len(component) == 0 {
//				continue
//			}
//			ch := p.GetChild(component)
//			if ch == nil {
//				// Create a directory
//				ch = p.NewPersistentInode(ctx, &mongonode.MongoNode{},
//					fs.StableAttr{Mode: syscall.S_IFDIR})
//				// Add it
//				p.AddChild(component, ch, true)
//			}
//
//			p = ch
//		}
//
//		// Make a file out of the content bytes. This type
//		// provides the open/read/flush methods.
//		embedder := &mongonode.MongoNode{
//			Data:     []byte(""), // Make empty
//			Id:       eachFile.ID,
//			FilePath: eachFile.FileName,
//		}
//
//		// Create the file. The Inode must be persistent,
//		// because its life time is not under control of the
//		// kernel.
//		child := p.NewPersistentInode(ctx, embedder, fs.StableAttr{Mode: syscall.S_IFREG})
//
//		// And add it
//		p.AddChild(base, child, true)
//	}
//}

// This demonstrates how to build a file system in memory. The
// read/write logic for the file is provided by the MemRegularFile type.
func main() {
	// This is where we'lcd ..l mount the FS
	mntDir, _ := ioutil.TempDir("", "")

	root := &mongonode.MongoNode{
		FilePath: "/",
	}
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{Debug: false},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Mounted on %s", mntDir)
	log.Printf("Unmount by calling 'fusermount -u %s'", mntDir)

	// Wait until unmount before exiting
	server.Wait()
}

//package main
//
//import (
//	"context"
//	"fmt"
//	"github.com/hanwen/go-fuse/v2/fs"
//	"github.com/hanwen/go-fuse/v2/fuse"
//	"log"
//	"path/filepath"
//	"strings"
//	"sync"
//	"syscall"
//)
//
//// bytesFileHandle allows reads
////var _ = (fs.FileReader)((*bytesFileHandle)(nil))
////
////func (fh *bytesFileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
////	log.Printf("Offset: %d, length: %d, Content: %d\n", off, int64(len(dest)), len(fh.content))
////	end := off + int64(len(dest))
////	if end > int64(len(fh.content)) {
////		end = int64(len(fh.content))
////	}
////
////	// We could copy to the `dest` buffer, but since we have a
////	// []byte already, return that.
////	return fuse.ReadResultData(fh.content[off:end]), 0
////}
//
//// timeFile is a file that contains the wall clock time as ASCII.
//type timeFile struct {
//	fs.Inode
//
//	// Only Read
//	mu      sync.Mutex
//	content []byte
//	id      string
//}
//
//// timeFile implements Open
//var _ = (fs.NodeOpener)((*timeFile)(nil))
//
//func (f *timeFile) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
//	log.Printf("Open called!\n")
//
//	// disallow writes(R/O)
//	if fuseFlags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
//		return nil, 0, syscall.EROFS
//	}
//
//	// Return FOPEN_DIRECT_IO so content is not cached.
//	return nil, 0, 0
//}
//
//// timeFile implements Read
//var _ = (fs.NodeReader)((*timeFile)(nil))
//
//func (f *timeFile) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
//	log.Printf("Content Size: %d, Offset: %d\n", len(f.content), off)
//	f.mu.Lock()
//	defer f.mu.Unlock()
//
//	endOffset := off + int64(len(dest))
//
//	if endOffset > int64(len(f.content)) {
//		// End offset is too large. Fit to content.
//		endOffset = int64(len(f.content))
//	}
//
//	return fuse.ReadResultData(f.content[off:endOffset]), 0
//}
//
//// Ensure that we implement NodeOnAdder
//var _ = (fs.NodeOnAdder)((*timeFile)(nil))
//
//// OnAdd is called on mounting the file system. Use it to populate
//// the file system tree.
//func (f *timeFile) OnAdd(ctx context.Context) {
//	fileList := GetAllFileMetadata()
//	log.Printf("Got: %d of metadatas\n", len(fileList))
//
//	// So for all metadata
//	for _, eachFile := range fileList {
//		// List directory and base(i.e "test/sub/another/test.txt"'s dir would be "test/sub/another" and base would be "test.txt"(
//		dir, base := filepath.Split(eachFile.FileName)
//
//		rootNodePtr := &f.Inode
//
//		// Do Subdirectory
//		for _, component := range strings.Split(dir, "/") {
//			if len(component) == 0 {
//				continue //No-OP
//			}
//
//			childPtr := rootNodePtr.GetChild(component)
//
//			if childPtr == nil {
//				// Directory does not exists.
//				childPtr = rootNodePtr.NewPersistentInode(ctx, &fs.Inode{}, fs.StableAttr{Mode: syscall.S_IFDIR})
//				rootNodePtr.AddChild(component, childPtr, true)
//			}
//
//			rootNodePtr = childPtr
//		}
//
//		embedder := &timeFile{
//			content: []byte(""),
//		}
//
//		child := rootNodePtr.NewPersistentInode(ctx, embedder, fs.StableAttr{})
//		log.Println("sub")
//		rootNodePtr.AddChild(base, child, true)
//	}
//
//	log.Printf("fin\n")
//}
//
//// ExampleDirectIO shows how to create a file whose contents change on
//// every read.
//func main() {
//	//GetFileMetadata("testFile")
//	mntDir := "/tmp/x"
//	root := &timeFile{}
//
//	// Mount the file system
//	server, err := fs.Mount(mntDir, root, &fs.Options{
//		MountOptions: fuse.MountOptions{Debug: true},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("cat %s/clock to see the time\n", mntDir)
//	fmt.Printf("Unmount by calling 'fusermount -u %s'\n", mntDir)
//
//	// Serve the file system, until unmounted by calling fusermount -u
//	server.Wait()
//}
