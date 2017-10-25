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
    domainInfo, err := domain.GetInfo()
    if err != nil {
      return err
    }

    uuid, err := domain.GetUUIDString()
    if err != nil {
      return err
    }
    fields := map[string]interface{}{
      "vm": uuid,
    }
    tags := make(map[string]string)
    fields["cpu_time"] = domainInfo.CpuTime
    fields["nr_virt_cpu"] = domainInfo.NrVirtCpu

    stats, err := domain.MemoryStats(10,0)
    if err != nil {
      return err
    }
    m := map[int32]string{
      int32(lv.DOMAIN_MEMORY_STAT_AVAILABLE): "mem_max",
      int32(lv.DOMAIN_MEMORY_STAT_USABLE): "mem_free",
    }
    for _, stat := range stats {
      if val, ok := m[stat.Tag]; ok {
        fields[val] = stat.Val
      }
    }
    acc.AddFields("vm.data", fields, tags)

    GatherInterfaces(*connection, domain, acc, uuid)

    GatherDisks(*connection, domain, acc, uuid)

    domain.Free()
  }

  return nil
}

func GatherInterfaces(c lv.Connect, d lv.Domain, acc telegraf.Accumulator, uuid string) error {
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
    fields := map[string]interface{}{
      "vm"         : uuid,
      "rx_bytes"   : iface.RxBytes,
      "rx_packets" : iface.RxPkts,
      "rx_errs"    : iface.RxErrs,
      "rx_drop"    : iface.RxDrop,
      "tx_bytes"   : iface.TxBytes,
      "tx_packets" : iface.TxPkts,
      "tx_errs"    : iface.TxErrs,
      "tx_drop"    : iface.TxDrop,
    }
    acc.AddFields("vm.data", fields, map[string]string{"interface": iface.Name})
  }
  return nil
}

func GatherDisks(c lv.Connect, d lv.Domain, acc telegraf.Accumulator, uuid string) error {
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
    fields := map[string]interface{}{
      "vm"       : uuid,
      "rd_req"   : disk.RdReqs,
      "rd_bytes" : disk.RdBytes,
      "wr_req"   : disk.WrReqs,
      "wr_bytes" : disk.WrBytes,
      "errs"     : disk.Errors,
    }
    acc.AddFields("vm.data", fields, map[string]string{"disk": disk.Name})
  }
  return nil
}

func init() {
  inputs.Add("libvirt", func() telegraf.Input {
    return &Libvirt{}
  })
}
