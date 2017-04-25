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
    domainName, err := domain.GetName()
    if err != nil {
      return err
    }

    domainInfo, err := domain.GetInfo()
    if err != nil {
      return err
    }

		tags := map[string]string{"domain": domainName}
    acc.AddFields("cpu_time", map[string]interface{}{"value": domainInfo.CpuTime} , tags)
    acc.AddFields("max_mem", map[string]interface{}{"value": domainInfo.MaxMem}, tags)
    acc.AddFields("memory", map[string]interface{}{"value": domainInfo.Memory}, tags)
    acc.AddFields("nr_virt_cpu", map[string]interface{}{"value": domainInfo.NrVirtCpu}, tags)

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
  for _, iface := range domStat[0].Net {
    tags["Name"] = iface.Name
    acc.AddFields("interface.RxBytes", map[string]interface{}{"value": iface.RxBytes}, tags)
    acc.AddFields("interface.RxPkts", map[string]interface{}{"value": iface.RxPkts}, tags)
    acc.AddFields("interface.RxErrs", map[string]interface{}{"value": iface.RxErrs}, tags)
    acc.AddFields("interface.RxDrop", map[string]interface{}{"value": iface.RxDrop}, tags)
    acc.AddFields("interface.TxBytes", map[string]interface{}{"value": iface.TxBytes}, tags)
    acc.AddFields("interface.TxPkts", map[string]interface{}{"value": iface.TxPkts}, tags)
    acc.AddFields("interface.TxErrs", map[string]interface{}{"value": iface.TxErrs}, tags)
    acc.AddFields("interface.TxDrop", map[string]interface{}{"value": iface.TxDrop}, tags)
    delete(tags, "Name")
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
  for _, disk := range domStats[0].Block {
    tags["Name"] = disk.Name
    acc.AddFields("disk.RdReqs", map[string]interface{}{"value": disk.RdReqs}, tags)
    acc.AddFields("disk.RdBytes", map[string]interface{}{"value": disk.RdBytes}, tags)
    acc.AddFields("disk.WrReqs", map[string]interface{}{"value": disk.WrReqs}, tags)
    acc.AddFields("disk.WrBytes", map[string]interface{}{"value": disk.WrBytes}, tags)
    acc.AddFields("disk.Errors", map[string]interface{}{"value": disk.Errors}, tags)
    delete(tags, "Name")
  }
  return nil
}


func init() {
	inputs.Add("libvirt", func() telegraf.Input {
		return &Libvirt{}
	})
}
