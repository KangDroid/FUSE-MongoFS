package mongonode

import (
	"context"
	"encoding/base64"
	"filesystem/mongocom"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"path/filepath"
	"sync"
	"syscall"
)

type MongoNode struct {
	fs.Inode

	mu       sync.Mutex
	Data     []byte
	Id       primitive.ObjectID
	FilePath string
}

var _ = (fs.NodeOpener)((*MongoNode)(nil))
var _ = (fs.NodeReader)((*MongoNode)(nil))
var _ = (fs.NodeGetattrer)((*MongoNode)(nil))
var _ = (fs.NodeSetattrer)((*MongoNode)(nil))
var _ = (fs.NodeFsyncer)((*MongoNode)(nil))
var _ = (fs.NodeFlusher)((*MongoNode)(nil))
var _ = (fs.NodeCreater)((*MongoNode)(nil))
var _ = (fs.NodeUnlinker)((*MongoNode)(nil))

//var _ = (fs.NodeRenamer)((*MongoNode)(nil))

//func (f *MongoNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
//	log.Printf("Name: %s, New Name: %s\n", name, newName)
//
//	dir, _ := filepath.Split(f.FilePath)
//
//	f.FilePath = filepath.Join(dir, newName)
//	return 0
//}

// Folder Related(EXP)
var _ = (fs.NodeOpendirer)((*MongoNode)(nil))
var _ = (fs.NodeRmdirer)((*MongoNode)(nil))
var _ = (fs.NodeMkdirer)((*MongoNode)(nil))

func (f *MongoNode) Unlink(ctx context.Context, name string) syscall.Errno {
	log.Printf("Trying to unlink: %s AND Name: %s\n", f.FilePath, name)

	mongocom.RemoveFile(f.FilePath, name)
	return 0
}

func (f *MongoNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	fullBasePath := f.FilePath

	log.Printf("Base Path: %s\n", fullBasePath)
	log.Printf("FilePath(Create): %s\n", fmt.Sprintf("%s/%s", fullBasePath, name))

	embedder := &MongoNode{
		Data:     []byte(base64.StdEncoding.EncodeToString([]byte(""))),
		FilePath: fmt.Sprintf("%s/%s", fullBasePath, name),
	}

	inodePtr := f.Inode.NewPersistentInode(ctx, embedder, fs.StableAttr{})

	return inodePtr, nil, 0, 0
}

func (f *MongoNode) Fsync(ctx context.Context, fh fs.FileHandle, flags uint32) syscall.Errno {
	log.Printf("Fsync Called: %s\n", string(f.Data))

	dir, base := filepath.Split(f.FilePath)

	mongocom.WriteFile(&mongocom.FileStruct{
		ID:          f.Id,
		FileParent:  dir,
		FileName:    base,
		FileContent: string(f.Data),
	})

	return 0
}

func (f *MongoNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	out.Attr.Size = uint64(len(f.Data))
	out.Attr.Mode = 0777
	return fs.OK
}

func (f *MongoNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	f.mu.Lock()
	defer f.mu.Unlock()
	if sz, ok := in.GetSize(); ok {
		f.Data = f.Data[:sz]
	}
	out.Size = uint64(len(f.Data))
	out.Attr.Mode = 0777
	return fs.OK
}
func (f *MongoNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// Disable READ OP
	//if fuseFlags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
	//	return nil, 0, syscall.EROFS
	//}

	log.Println("Open Called")
	log.Printf("File Name: %s\n", f.FilePath)

	// Find File Object
	mongoStruct := mongocom.FindFileByName(f.FilePath)
	encodedStr := mongoStruct.FileContent
	decodedArr, _ := base64.StdEncoding.DecodeString(encodedStr)

	f.Data = decodedArr

	return nil, fuse.FOPEN_DIRECT_IO, fs.OK
}

func (f *MongoNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	f.mu.Lock()
	defer f.mu.Unlock()

	end := int(off) + len(dest)
	if end > len(f.Data) {
		end = len(f.Data)
	}
	return fuse.ReadResultData(f.Data[off:end]), fs.OK
}

var _ = (fs.NodeWriter)((*MongoNode)(nil))

func (f *MongoNode) Write(ctx context.Context, fh fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	log.Printf("Write Called\n")
	f.mu.Lock()
	defer f.mu.Unlock()
	end := int64(len(data)) + off
	if int64(len(f.Data)) < end {
		n := make([]byte, end)
		copy(n, f.Data)
		f.Data = n
	}

	copy(f.Data[off:off+int64(len(data))], data)

	return uint32(len(data)), 0
}

func (f *MongoNode) Flush(ctx context.Context, fh fs.FileHandle) syscall.Errno {
	encodedData := base64.StdEncoding.EncodeToString(f.Data)
	log.Printf("Encoded String: %s\n", encodedData)

	dir, base := filepath.Split(f.FilePath)
	log.Printf("Flush Dir: %s, Base: %s\n", dir, base)

	mongocom.WriteFile(&mongocom.FileStruct{
		ID:          f.Id,
		FileParent:  dir,
		FileName:    base,
		FileContent: encodedData,
	})

	return 0
}

func (f *MongoNode) Opendir(ctx context.Context) syscall.Errno {
	fmt.Printf("Just opened DIR %s\n", f.FilePath)
	fmt.Printf("Dir Input: %s\n", f.FilePath)

	// Remove Tree
	f.Inode.RmAllChildren()

	listDir := mongocom.ListDirectory(f.FilePath)
	fmt.Printf("List Size: %d\n", len(listDir))

	for _, eachStruct := range listDir {
		if eachStruct.IsFolder {
			nodePtr := f.Inode.NewPersistentInode(ctx, &MongoNode{FilePath: filepath.Join(eachStruct.FileParent, eachStruct.FileName)}, fs.StableAttr{Mode: syscall.S_IFDIR})
			f.Inode.AddChild(eachStruct.FileName, nodePtr, true)
		} else {
			nodePtr := f.Inode.NewPersistentInode(ctx, &MongoNode{FilePath: filepath.Join(eachStruct.FileParent, eachStruct.FileName)}, fs.StableAttr{Mode: syscall.S_IFREG})
			f.Inode.AddChild(eachStruct.FileName, nodePtr, true)
		}
	}

	return 0
}

func (f *MongoNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Printf("Making directory is not supported yet!\n")
	log.Printf("Target Directory: %s, name: %s\n", f.FilePath, name)

	//nodePtr := f.Inode.NewPersistentInode(ctx, &MongoNode{
	//	FilePath: filepath.Join(f.FilePath, name),
	//}, fs.StableAttr{Mode: syscall.S_IFDIR})
	//f.Inode.AddChild(name, nodePtr, true)
	//
	//mongocom.CreateDirectory(f.FilePath, name)

	return nil, syscall.EOPNOTSUPP
}

func (f *MongoNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	log.Printf("Trying to remove directory: %s, Name: %s\n")

	return syscall.EOPNOTSUPP
}
