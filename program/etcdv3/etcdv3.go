package etcdv3

import (
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/qiuhoude/etcd-manage/program/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// EtcdClis etcd连接对象
	etcdClis *sync.Map
)

func init() {
	etcdClis = new(sync.Map)
}

// Etcd3Client etcd v3客户端
type Etcd3Client struct {
	*clientv3.Client
}

//  NewEtcdCli 创建一个etcd客户端
func NewEtcdCli(etcdCfg *config.EtcdServer) (*Etcd3Client, error) {
	// 配置检测
	if etcdCfg == nil {
		return nil, errors.New("etcdCfg is nil")
	}
	if etcdCfg.TLSEnable && etcdCfg.TLSConfig == nil {
		return nil, errors.New("TLSConfig is nil")
	}
	if len(etcdCfg.Address) == 0 {
		return nil, errors.New("Etcd connection address cannot be empty")
	}
	// etcd 需要的配置
	cliCfg := clientv3.Config{
		Endpoints:   etcdCfg.Address,
		DialTimeout: 10 * time.Second,
		Username:    etcdCfg.Username,
		Password:    etcdCfg.Password,
	}

	if etcdCfg.TLSEnable { // 开启tls
		tlsInfo := transport.TLSInfo{
			CertFile:      etcdCfg.TLSConfig.CertFile,
			KeyFile:       etcdCfg.TLSConfig.KeyFile,
			TrustedCAFile: etcdCfg.TLSConfig.CAFile,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, err
		}
		cliCfg.TLS = tlsConfig
	}
	cli, err := clientv3.New(cliCfg)
	if err != nil {
		return nil, err
	}
	// 保存到map中
	etcdClis.Store(etcdCfg.Name, cli)
	return &Etcd3Client{cli}, nil
}

// GetEtcdCli 获取一个etcd cli对象
func GetEtcdCli(etcdCfg *config.EtcdServer) (*Etcd3Client, error) {
	if etcdCfg == nil {
		return nil, errors.New("etcdCfg is nil")
	}
	// 从sync.map中获取
	val, ok := etcdClis.Load(etcdCfg.Name)
	if !ok {
		if len(etcdCfg.Address) > 0 { // 没有取到就去创建
			cli, err := NewEtcdCli(etcdCfg)
			if err != nil {
				return nil, err
			}
			return cli, nil
		}
		return nil, errors.New("Getting etcd client error")
	}
	return &Etcd3Client{val.(*clientv3.Client)}, nil
}

// node 列表格式化成json
func NodeJsonFormat(prefix string, list []*Node) (interface{}, error) {
	ret := make(map[string]interface{}, 0)
	if len(list) == 0 {
		return ret, nil
	}
	for _, n := range list {
		key := strings.TrimPrefix(n.FullDir, prefix)
		key = strings.TrimRight(key, "/")
		strs := strings.Split(key, "/")
		recursiveJsonMap(strs, n, ret)
	}
	return ret, nil
}

//递归的将一个值赋值到map中
func recursiveJsonMap(strs []string, node *Node, parent map[string]interface{}) {
	if len(strs) == 0 || strs[0] == "" || node == nil || parent == nil { // 递归结束条件
		return
	}
	if _, ok := parent[strs[0]]; !ok { // 不存在创建目录
		if node.Value == DEFAULT_DIR_VALUE {
			parent[strs[0]] = make(map[string]interface{}, 0)
		} else {
			parent[strs[0]] = formatValue(node.Value)
		}
	}

	if val, ok := parent[strs[0]].(map[string]interface{}); ok {
		recursiveJsonMap(strs[1:], node, val)
	}
}

// Format 时获取值，转为指定类型
func formatValue(v string) interface{} {
	if strings.EqualFold(v, "true") {
		return true
	} else if strings.EqualFold(v, "false") {
		return false
	}
	// 尝试转浮点数
	vf, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return vf
	}
	// 尝试转整数
	vi, err := strconv.ParseInt(v, 10, 64)
	if err == nil {
		return vi
	}
	return v
}
