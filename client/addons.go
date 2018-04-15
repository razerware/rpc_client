package client

import (
	"net"
	"github.com/golang/glog"
	"fmt"
	"os"
)

func GetInternal() (int, string, string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		glog.Fatal("Oops:" + err.Error())
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//ipnet.IP.String()
				sql := fmt.Sprintf("SELECT * FROM `vm_info` where `ip`='%s'", ipnet.IP.String())
				record := MysqlQuery(sql)
				if len(record) > 0 {
					v1, _ := record[0]["inner_id"].(int)
					v2, _ := record[0]["ip"].(string)
					v3, _ := record[0]["swarm_id"].(string)
					glog.Info("Hostid is :",v1,"IP is: " ,v2, "Swarm_id is:",v3)
					return v1, v2, v3
				}
			}
		}
	}
	glog.Fatal("ip get error")
	os.Exit(0)
	return 0, "", ""
}