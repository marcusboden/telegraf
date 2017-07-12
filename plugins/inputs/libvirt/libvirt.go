package libvirt

import (
	lv "github.com/libvirt/libvirt-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  # specify a libvirt connection uri
  uri = "qemu:///system"
`

type Libvirt struct {
	Uri string
}

func (l *Libvirt) SampleConfig() string {
	return sampleConfig
}

func (l *Libvirt) Description() string {
	return "Read domain infos from a libvirt deamon"
}

func (l *Libvirt) Gather(acc telegraf.Accumulator) error {
	connection, err := lv.NewConnectReadOnly(l.Uri)
	if err != nil {
		return err
	}
	defer connection.Close()

	domains, err := connection.ListAllDomains(lv.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return err
	}

	for _, domain := range domains {
    defer domain.Free()
    domainName, err := domain.GetName()
    if err != nil {
      return err
    }

    domainInfo, err := domain.GetInfo()
    if err != nil {
      return err
    }

    tags := map[string]string{"domain": domainName, "cloud": "new"}
    acc.AddFields("vm.cpu_time", map[string]interface{}{"value": float64(domainInfo.CpuTime)} , tags)
    acc.AddFields("vm.max_mem", map[string]interface{}{"value": float64(domainInfo.MaxMem)}, tags)
    acc.AddFields("vm.memory", map[string]interface{}{"value": float64(domainInfo.Memory)}, tags)
    acc.AddFields("vm.nr_virt_cpu", map[string]interface{}{"value": float64(domainInfo.NrVirtCpu)}, tags)

    GatherInterfaces(*connection, domain, acc, tags)

    GatherDisks(*connection, domain, acc, tags)
	}

	return nil
}

func GatherInterfaces(c lv.Connect, d lv.Domain, acc telegraf.Accumulator , tags map[string]string) error {
  domStat, err := c.GetAllDomainStats(
    []*lv.Domain{&d},
    lv.DOMAIN_STATS_INTERFACE,
    lv.CONNECT_GET_ALL_DOMAINS_STATS_ACTIVE,
  )
  if err != nil {
    return err
  }
  defer domStat[0].Domain.Free()
  for _, iface := range domStat[0].Net {
    tags["name"] = iface.Name
    acc.AddFields("vm.interface.rx_bytes", map[string]interface{}{"value": float64(iface.RxBytes)}, tags)
    acc.AddFields("vm.interface.rx_packets", map[string]interface{}{"value": float64(iface.RxPkts)}, tags)
    acc.AddFields("vm.interface.rx_errs", map[string]interface{}{"value": float64(iface.RxErrs)}, tags)
    acc.AddFields("vm.interface.rx_drop", map[string]interface{}{"value": float64(iface.RxDrop)}, tags)
    acc.AddFields("vm.interface.tx_bytes", map[string]interface{}{"value": float64(iface.TxBytes)}, tags)
    acc.AddFields("vm.interface.tx_packets", map[string]interface{}{"value": float64(iface.TxPkts)}, tags)
    acc.AddFields("vm.interface.tx_errs", map[string]interface{}{"value": float64(iface.TxErrs)}, tags)
    acc.AddFields("vm.interface.tx_drop", map[string]interface{}{"value": float64(iface.TxDrop)}, tags)
    delete(tags, "name")
  }
  return nil
}

func GatherDisks(c lv.Connect, d lv.Domain, acc telegraf.Accumulator , tags map[string]string) error {
  domStats, err := c.GetAllDomainStats(
    []*lv.Domain{&d},
    lv.DOMAIN_STATS_BLOCK,
    lv.CONNECT_GET_ALL_DOMAINS_STATS_ACTIVE,
  )
  if err != nil {
    return err
  }
  defer domStats[0].Domain.Free()
  for _, disk := range domStats[0].Block {
    tags["name"] = disk.Name
    acc.AddFields("vm.disk.rd_req", map[string]interface{}{"value": float64(disk.RdReqs)}, tags)
    acc.AddFields("vm.disk.rd_bytes", map[string]interface{}{"value": float64(disk.RdBytes)}, tags)
    acc.AddFields("vm.disk.wr_req", map[string]interface{}{"value": float64(disk.WrReqs)}, tags)
    acc.AddFields("vm.disk.wr_bytes", map[string]interface{}{"value": float64(disk.WrBytes)}, tags)
    acc.AddFields("vm.disk.errs", map[string]interface{}{"value": float64(disk.Errors)}, tags)
    delete(tags, "name")
  }
  return nil
}


func init() {
	inputs.Add("libvirt", func() telegraf.Input {
		return &Libvirt{}
	})
}
