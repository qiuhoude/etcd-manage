package program

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/qiuhoude/etcd-manage/program/config"
	"github.com/qiuhoude/etcd-manage/program/etcdv3"
	"github.com/qiuhoude/etcd-manage/program/v1"
	"log"
	"net/http"
	"strings"
	"time"
)

// 启动http服务
func (p *Program) startAPI() {
	router := gin.Default()
	//设置跨域中间件
	router.Use(p.middleware())

	// 设置静态文件目录
	router.GET("/ui/*w", p.handlerStatic)
	// 重定向到静态文件中
	router.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/ui")
	})

	// 读取认证列表,添加到gin.Accounts中
	accounts := make(gin.Accounts, 0)
	if p.cfg.Users != nil {
		for _, u := range p.cfg.Users {
			accounts[u.Username] = u.Password
		}
	}
	// v1 api
	apiV1 := router.Group("/v1", gin.BasicAuth(accounts))
	apiV1.Use(p.middlewareEtcd()) // 绑定etcd客户端中间件
	v1.V1(apiV1)

	// 启动http服务
	addr := fmt.Sprintf("%s:%d", p.cfg.HTTP.Address, p.cfg.HTTP.Port)
	s := http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("启动HTTP服务:", addr)
	// TLS 判断
	var err error
	if p.cfg.HTTP.TLSEnable {
		if p.cfg.HTTP.TLSConfig == nil || p.cfg.HTTP.TLSConfig.CertFile == "" || p.cfg.HTTP.TLSConfig.KeyFile == "" {
			log.Fatalln("启用tls必须配置证书文件路径")
		}
		err = s.ListenAndServeTLS(p.cfg.HTTP.TLSConfig.CertFile, p.cfg.HTTP.TLSConfig.KeyFile)
	} else if p.cfg.HTTP.TLSEncryptEnable {
		if len(p.cfg.HTTP.TLSEncryptDomainNames) == 0 {
			log.Fatalln("域名列表不能为空")
		}
		err = autotls.Run(router, p.cfg.HTTP.TLSEncryptDomainNames...)
	} else {
		err = s.ListenAndServe()
	}

	if err != nil {
		log.Fatalln(err)
	}
}

// --------------------------- 中间件 -------------------------
// 跨域中间件
func (p *Program) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// gin设置响应头，设置跨域
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Access-Control-Allow-Origin")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		//放行所有OPTIONS方法
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
		}
	}
}

// etcd客户端中间件
func (p *Program) middlewareEtcd() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取登录用户名，查询角色
		userIn := c.MustGet(gin.AuthUserKey)
		userRole := ""
		if userIn != nil {
			user := userIn.(string)
			if user == "" {
				c.Set("userRole", "")
			} else {
				u := p.cfg.GetUserByUsername(user)
				if u == nil {
					c.Set("userRole", "")
				} else {
					userRole = u.Role
					// 角色和用户信息
					c.Set("userRole", u.Role)
					c.Set("authUser", u)
				}
			}
		}

		// 绑定etcd 连接
		etcdServerName := c.GetHeader("EtcdServerName")
		fmt.Println("etcdServerName ->", etcdServerName)
		if strings.EqualFold("", etcdServerName) || strings.EqualFold("null", etcdServerName) {
			etcdServerName = "default"
		}
		if etcdServerName != "" {
			cli, s, err := getEtcdCli(etcdServerName, userRole)
			if err == nil {
				c.Set("EtcdServer", cli)
				c.Set("EtcdServerCfg", s)
			}
		}
		c.Next()
	}
}

func getEtcdCli(etcdServerName, userRole string) (cli *etcdv3.Etcd3Client, s *config.EtcdServer, err error) {
	s = config.GetEtcdServer(etcdServerName)
	if s == nil {
		return nil, nil, errors.New("etcd服务不存在")
	}
	// 查看允许访问的角色
	if len(s.Roles) > 0 {
		isRole := false
		for _, r := range s.Roles {
			if r == userRole {
				isRole = true
				break
			}
		}
		if !isRole {
			return nil, nil, errors.New("无权限访问")
		}
	}
	cli, err = etcdv3.GetEtcdCli(s)
	return
}
