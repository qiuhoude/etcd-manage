package etcdv3

import (
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/qiuhoude/etcd-manage/program/config"
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
