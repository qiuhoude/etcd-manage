package etcdv3

import (
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"strings"
)

const (
	ROLE_LEADER   = "leader"
	ROLE_FOLLOWER = "follower"

	STATUS_HEALTHY   = "healthy"
	STATUS_UNHEALTHY = "unhealthy"

	// 目录的默认值
	DEFAULT_DIR_VALUE = "etcdv3_dir_$2H#%gRe3*t"
)

// Member 节点信息
type Member struct {
	*etcdserverpb.Member
	Role   string `json:"role"`
	Status string `json:"status"`
	DbSize int64  `json:"db_size"`
}

// Node 需要使用到的模型
type Node struct {
	IsDir   bool   `json:"is_dir"`
	Version int64  `json:"version,string"`
	Value   string `json:"value"`
	FullDir string `json:"full_dir"`
}

func NewNode(dir string, kv *mvccpb.KeyValue) *Node {
	return &Node{
		IsDir:   string(kv.Value) == DEFAULT_DIR_VALUE,
		Version: kv.Version,
		Value:   strings.TrimPrefix(string(kv.Key), dir),
		FullDir: string(kv.Key),
	}
}
