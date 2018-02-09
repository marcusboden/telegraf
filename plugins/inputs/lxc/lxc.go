package lxc

// lxc.go

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type Lxc struct {
	Metrics_whitelist   []string
	Metrics_blacklist   []string
	Container_whitelist []string
	Container_blacklist []string
}

const Config string = `
# Metrics: By default, all metrics are collected.
# If metrics_whitelist is specified, only those metrics are collected.
# If metrics_blacklist is specified, those metrics are left out.
# These options are mutually exclusive.
metrics_whitelist = []
metrics_blacklist = []

# Containers: By default, metrics of all containers are collected.
# A blacklist or a whitelist can be specified, but similar to the metrics-
# options, both are mutually exclusive.
container_whitelist = []
container_blacklist = []
`
const BaseString string = "/sys/fs/cgroup/%s/lxc"

func GatherMemInfo(ContainerName string) (map[string]interface{}, error) {
	file := "memory.stat"
	mem := path.Join(fmt.Sprintf(BaseString, "memory"), ContainerName, file)
	dat, err := ioutil.ReadFile(mem)
	if err != nil {
		return nil, err
	}
	FileArray := strings.Split(strings.TrimSuffix(string(dat), "\n"), "\n")
	ContainerMap := make(map[string]interface{})
	for _, l := range FileArray {
		tmp := strings.Split(l, " ")
		ContainerMap[tmp[0]], err = strconv.Atoi(tmp[1])
		if err != nil {
			return nil, err
		}
	}
	return ContainerMap, nil
}

func GatherCpuInfo(ContainerName string) (map[string]interface{}, error) {
	file := "cpuacct.stat"
	cpu := path.Join(fmt.Sprintf(BaseString, "cpuacct"), ContainerName, file)
	dat, err := ioutil.ReadFile(cpu)
	if err != nil {
		return nil, err
	}
	FileArray := strings.Split(strings.TrimSuffix(string(dat), "\n"), "\n")
	ContainerMap := make(map[string]interface{})
	for _, l := range FileArray {
		tmp := strings.Split(l, " ")
		ContainerMap[tmp[0]], err = strconv.Atoi(tmp[1])
		if err != nil {
			return nil, err
		}
	}
	return ContainerMap, nil
}

func GatherBlockIO(ContainerName string) (map[string]interface{}, error) {
	var BlkioFiles []string
	for _, v := range [4]string{"blkio.sectors", "blkio.io_service_bytes", "blkio.io_serviced", "blkio.io_queued"} {
		BlkioFiles = append(BlkioFiles, v, v+"_recursive")
	}
	ContainerMap := make(map[string]interface{})
	for _, f := range BlkioFiles {
		blk := path.Join(fmt.Sprintf(BaseString, "blkio"), ContainerName, f)
		dat, err := ioutil.ReadFile(blk)
		if err != nil {
			return nil, err
		}
		if len(dat) > 0 {
			FileArray := strings.Split(strings.TrimSuffix(string(dat), "\n"), "\n")
			for _, l := range FileArray {
				tmp := strings.Split(l, " ")
				ContainerMap[f+tmp[0]], err = strconv.Atoi(tmp[1])
				if err != nil {
					return nil, err
				}
			}
		} else {
			ContainerMap[f] = 0
		}
	}
	return ContainerMap, nil
}

func (l *Lxc) Description() string {
	return "Gathers performance metrics of LXC Containers"
}

func (l *Lxc) SampleConfig() string {
	return Config
}

func isec(s1 []string, s2 []string) []string {
	set := make(map[string]bool)
	list := make([]string, 0)
	for _, k := range s2 {
		set[k] = true
	}
	for _, k := range s1 {
		if set[k] {
			list = append(list, k)
		}
	}
	return list
}

func diff(s1 []string, s2 []string) []string {
	set := make(map[string]bool)
	for _, k := range s1 {
		set[k] = true
	}
	for _, k := range s2 {
		set[k] = false
	}
	list := make([]string, 0)
	for k := range set {
		if set[k] {
			list = append(list, k)
		}
	}
	return list
}

func (l *Lxc) GetContainers() ([]string, error) {
	//get available lxc containers
	cmd := exec.Command("lxc", "list", "-c", "ns", "--format", "csv")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Calling \"%s\" caused an error: %s.", strings.Join(cmd.Args, " "), err.Error())
	}
	//Trim last newline and split by newline
	var availContainers []string
	for _, v := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
		tmp := strings.Split(v, ",")
		if tmp[1] == "RUNNING" {
			availContainers = append(availContainers, tmp[0])
		}
	}

	if len(l.Container_whitelist) > 0 {
		if len(l.Container_blacklist) > 0 {
			return nil, errors.New("Containers blacklist and whitelist cannot both be declared")
		}
		return isec(availContainers, l.Container_whitelist), nil
	} else if len(l.Container_blacklist) > 0 {
		return diff(availContainers, l.Container_blacklist), nil
	}
	return availContainers, nil
}

func (l *Lxc) FilterMetrics(ContainerMap map[string]interface{}) (map[string]interface{}, error) {
	availMetrics := make([]string, 0)
	for k := range ContainerMap {
		availMetrics = append(availMetrics, k)
	}

	var filtered []string
	if len(l.Metrics_whitelist) > 0 {
		if len(l.Metrics_blacklist) > 0 {
			return nil, errors.New("Metrics blacklist and whitelist cannot both be declared")
		}
		filtered = isec(availMetrics, l.Metrics_whitelist)
	} else if len(l.Metrics_blacklist) > 0 {
		filtered = diff(availMetrics, l.Metrics_blacklist)
	} else {
		filtered = availMetrics
	}
	newMap := make(map[string]interface{})
	for _, m := range filtered {
		newMap[m] = ContainerMap[m]
	}
	return newMap, nil
}

func (l *Lxc) Gather(acc telegraf.Accumulator) error {
	var ContainerMap map[string]interface{}
	tag_name := "name"
	tag_cat := "category"
	containers, err := l.GetContainers()
	if err != nil {
		return err
	}
	for _, s := range containers {
		var err error

		ContainerMap, err = GatherMemInfo(s)
		if err != nil {
			return err
		}
		ContainerMap, err = l.FilterMetrics(ContainerMap)
		if err != nil {
			return err
		}
		acc.AddFields("lxc-container", ContainerMap, map[string]string{tag_name: s, tag_cat: "mem"})

		ContainerMap, err = GatherCpuInfo(s)
		if err != nil {
			return err
		}
		ContainerMap, err = l.FilterMetrics(ContainerMap)
		if err != nil {
			return err
		}
		acc.AddFields("lxc-container", ContainerMap, map[string]string{tag_name: s, tag_cat: "cpu"})

		ContainerMap, err = GatherBlockIO(s)
		if err != nil {
			return err
		}
		ContainerMap, err = l.FilterMetrics(ContainerMap)
		if err != nil {
			return err
		}
		acc.AddFields("lxc-container", ContainerMap, map[string]string{tag_name: s, tag_cat: "io"})
	}
	return nil
}

func init() {
	inputs.Add("lxc", func() telegraf.Input { return &Lxc{} })
}
