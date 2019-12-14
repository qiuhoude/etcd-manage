package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qiuhoude/etcd-manage/program/etcdv3"
	"net/http"
)

// V1 v1 版接口 路由入口
func V1(v1 *gin.RouterGroup) {
	v1.GET("/members", getEtcdMembers) // 获取节点列表
}

// 获取服务节点
func getEtcdMembers(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": err.Error(),
			})
		}
	}()

	etcdCli, exists := c.Get("EtcdServer")
	if exists == false {
		fmt.Println("-->Etcd client is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Etcd client is empty",
		})
		return
	}
	cli := etcdCli.(*etcdv3.Etcd3Client)

	members, err := cli.Members()
	fmt.Println("---->>>,len:", len(members))
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, members)
}
