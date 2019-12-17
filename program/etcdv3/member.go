package etcdv3

import (
	"context"
	"fmt"
	"time"
)

// Members 获取节点列表
func (c *Etcd3Client) Members() ([]*Member, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.Client.MemberList(ctx)
	if err != nil {
		return nil, err
	}
	members := make([]*Member, 0)
	for _, member := range resp.Members {
		if len(member.ClientURLs) > 0 {
			mm := member
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()
				m := &Member{
					Member: mm,
					Role:   ROLE_FOLLOWER,
					Status: STATUS_UNHEALTHY,
				}
				fmt.Println("url -> " + m.ClientURLs[0])
				resp, err := c.Client.Status(ctx, m.ClientURLs[0])
				if err == nil {
					m.Status = STATUS_HEALTHY
					m.DbSize = resp.DbSize
					if resp.Leader == resp.Header.MemberId {
						m.Role = ROLE_LEADER
					}
				}
				members = append(members, m)
			}()

		}
	}
	return members, nil

}
