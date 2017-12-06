package client

import (
	"log"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"strconv"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
)

var dbUrl = "http://10.109.252.172:8086"
var db = "test"

type HostInfo struct {
	Hostid int
	Hostip string
}
type containerMonitorStats struct {
	HostInfo
	CpuPercent  float64
	MemPercent  float64
	Name string
	serviceId string
	serviceName   string
}
type vmMonitorStats struct {
	HostInfo
	CpuPercent  float64
	MemPercent  float64
	swarmId int
}


type ContainerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage int64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			TotalUsage int64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage int   `json:"usage"`
		Limit int64 `json:"limit"`
	} `json:"memory_stats"`
}

type Container struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
	Labels struct {
		ComDockerSwarmServiceID   string `json:"com.docker.swarm.service.id"`
		ComDockerSwarmServiceName string `json:"com.docker.swarm.service.name"`
	} `json:"Labels"`
}

func CollectData(info HostInfo) {
	CollectVm(info)
	CollectContainer(info)
}

func CollectVm(info HostInfo)  {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	// almost every return value is a struct
	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total/1024/1024/1024, v.Free, v.UsedPercent)
	result := vmMonitorStats{info, c[0], v.UsedPercent, 1}
	// convert to JSON. String() is also implemented
	sendVmInfo("vm",result)
}
func getContainers(info HostInfo) []Container {
	client := &http.Client{}
	url:="http://"+info.Hostip+":2375/containers/json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle error
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	c := []Container{}
	json.Unmarshal(body, &c)
	return c
}
func CollectContainer(info HostInfo) {
	c := getContainers(info)
	for _, i := range c {
		if i.Labels.ComDockerSwarmServiceID == "" {
			continue
		} else {
			go func(i Container) {
				client := &http.Client{}
				url := "http://10.109.252.172:2375/containers/" + i.ID + "/stats?stream=false"
				fmt.Println(url)
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					// handle error
				}
				resp, err := client.Do(req)
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					// handle error
				}
				cs := ContainerStats{}
				json.Unmarshal(body, &cs)
				fmt.Println(cs)

				cpu_percent := float64((cs.CPUStats.CPUUsage.TotalUsage -
					cs.PrecpuStats.CPUUsage.TotalUsage)) /
					float64((cs.CPUStats.SystemCPUUsage - cs.PrecpuStats.SystemCPUUsage)) * 100
				mem_percent := float64(cs.MemoryStats.Usage) / float64(cs.MemoryStats.Limit) * 100
				service_id := i.Labels.ComDockerSwarmServiceID
				service_name := i.Labels.ComDockerSwarmServiceName
				ms := containerMonitorStats{info, cpu_percent, mem_percent, i.Names[0],
				service_id, service_name}
				sendContainerInfo("container", ms)
				fmt.Println(ms)

			}(i)

		}

	}
	return
}
func sendContainerInfo(field string, stat containerMonitorStats) {
	url := dbUrl + "/write?db=" + db
	tags := "hostid=" + strconv.Itoa(stat.Hostid) + ",serviceid=" + stat.serviceId + ",servicename=" +
		stat.serviceName + ",name="+stat.Name
	stat_string := field + "," + tags + " cpu=" + strconv.FormatFloat(stat.CpuPercent, 'f', 2, 64) +
		",mem=" + strconv.FormatFloat(stat.MemPercent, 'f', 2, 64)
	stat_byte := []byte(stat_string)
	fmt.Println(stat_string)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(stat_byte))
	if err != nil {
		// handle error
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
	} else {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	} else {
		log.Println(string(body))
	}
}
func sendVmInfo(field string, stat vmMonitorStats) {
	url := dbUrl + "/write?db=" + db
	tags := "hostid=" + strconv.Itoa(stat.Hostid) + ",swarmid=" + strconv.Itoa(stat.swarmId)
	stat_string := field + "," + tags + " cpu=" + strconv.FormatFloat(stat.CpuPercent, 'f', 2, 64) +
		",mem=" + strconv.FormatFloat(stat.MemPercent, 'f', 2, 64)
	stat_byte := []byte(stat_string)
	fmt.Println(stat_string)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(stat_byte))
	if err != nil {
		// handle error
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
	} else {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	} else {
		log.Println(string(body))
	}
}