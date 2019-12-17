package v1

import "github.com/qiuhoude/etcd-manage/program/etcdv3"

// PostReq 添加和修改时的body
type PostReq struct {
	*etcdv3.Node
	EtcdName string `json:"etcd_name"`
}

//日志信息
type LogLine struct {
	Date  string  `json:"date"`
	User  string  `json:"user"`
	Role  string  `json:"role"`
	Msg   string  `json:"msg"`
	Ts    float64 `json:"ts"`
	Level string  `json:"level"`
}
