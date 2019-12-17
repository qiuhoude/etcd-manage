package etcdv3

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"path"
	"strings"
	"time"
)

// 递归获取key
func (c *Etcd3Client) GetRecursiveValue(key string) (list []*Node, err error) {
	list = make([]*Node, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := c.Client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	// 过滤掉目录
	for _, kv := range resp.Kvs {
		list = append(list, &Node{
			Value:   string(kv.Value),
			FullDir: string(kv.Key),
			Version: kv.Version,
		})
	}
	//for _,node:=range list{
	//	fmt.Printf("node:%v\n",node)
	//}
	return
}

// 删除key
func (c *Etcd3Client) Delete(key string) error {
	key = strings.TrimRight(key, "/")
	dir := key + "/"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txn := c.Client.Txn(ctx)
	// 如果是目录就删除整个目录
	_, err := txn.If(
		clientv3.Compare(clientv3.Value(dir), "=", DEFAULT_DIR_VALUE),
	).Then(
		clientv3.OpDelete(key),
		clientv3.OpDelete(dir, clientv3.WithPrefix()), // 删除以dir目录未前缀的key
	).Else(
		clientv3.OpDelete(key), //非目录值删除当前key
	).Commit()
	return err
}

// 通过key 获取value
func (c *Etcd3Client) Value(key string) (val *Node, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.Client.Get(ctx, key)
	if err != nil {
		return
	}
	if resp.Kvs != nil && len(resp.Kvs) > 0 {
		val = &Node{
			Value:   string(resp.Kvs[0].Value),
			FullDir: key,
			Version: resp.Kvs[0].Version,
		}
	} else {
		err = ErrorKeyNotFound
	}
	return
}

// 返回key以及父路径
func (c *Etcd3Client) ensureKey(key string) (string, string) {
	key = strings.TrimRight(key, "/") // 去掉右边的 / , 比如 /etc/java/ 变成 /etc/java
	if key == "" { // 更目录
		return "/", ""
	}
	if strings.Contains(key, "/") {
		return key, path.Clean(key + "/../")
	} else {
		return key, ""
	}

}

// Put 添加一个key
func (c *Etcd3Client) Put(key string, value string, mustEmpty bool) error {

	key, parentKey := c.ensureKey(key)
	//  需要判断的条件
	cmp := make([]clientv3.Cmp, 0)

	if parentKey != "" { // 有父节点
		c := clientv3.Compare(clientv3.Value(parentKey), "=", DEFAULT_DIR_VALUE)
		cmp = append(cmp, c)
	}
	if mustEmpty {
		//c := clientv3.Compare(clientv3.Value(key), "=", "")
		//cmp = append(cmp, c)
	} else {
		c := clientv3.Compare(clientv3.Value(key), "!=", DEFAULT_DIR_VALUE)
		cmp = append(cmp, c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 创建事物
	txn := c.Client.Txn(ctx)
	txn.If( // 条件判断
		cmp...
	).Then( // 事物操作
		clientv3.OpPut(key, value),
	)
	// 提交事物
	txnResp, err := txn.Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded { // 添加失败
		return ErrorPutKey
	}
	return nil
}

func (c *Etcd3Client) List(key string) (nodes []*Node, err error) {
	nodes = make([]*Node, 0)
	if key == "" {
		return nodes, errors.New("key is empty")
	}
	// 兼容key前缀设置为 /
	dir := key
	if key != "/" {
		key = strings.TrimRight(key, "/")
		dir = key + "/"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 创建事物
	txn := c.Client.Txn(ctx)
	txnResp, err := txn.If( // 条件判断
		//clientv3.Compare(clientv3.Value(key), "=", DEFAULT_DIR_VALUE),
	).Then( // 事物操作
		clientv3.OpGet(dir, clientv3.WithPrefix()),
	).Commit()

	if err != nil {
		return nil, err
	}

	if txnResp.Succeeded {
		if len(txnResp.Responses) > 0 {
			rangeResp := txnResp.Responses[0].GetResponseRange()
			return c.list(dir, rangeResp.Kvs)
		} else {
			// empty directory
			return []*Node{}, nil
		}
	} else {
		return nil, ErrorListKey
	}
}

func (c *Etcd3Client) list(dir string, kvs []*mvccpb.KeyValue) ([]*Node, error) {
	nodes := make([]*Node, 0)
	for _, kv := range kvs {
		name := strings.TrimPrefix(string(kv.Key), dir)
		if strings.Contains(name, "/") {
			// secondary directory
			continue
		}
		nodes = append(nodes, NewNode(dir, kv))
	}
	return nodes, nil
}
