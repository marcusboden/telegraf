package libvirt

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	lv "github.com/libvirt/libvirt-go"
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

	acc.AddFields("vm.data", map[string]interface{}{"count": len(domains)}, make(map[string]string))

	for _, domain := range domains {
		domainInfo, err := domain.GetInfo()
		if err != nil {
			return err
		}

		uuid, err := domain.GetUUIDString()
		if err != nil {
			return err
		}

		tags := map[string]string{
			"uuid": uuid,
		}

		fields := map[string]interface{}{
			"cpu_time":    domainInfo.CpuTime,
			"nr_virt_cpu": domainInfo.NrVirtCpu,
		}
		stats, err := domain.MemoryStats(20, 0)
		if err != nil {
			return err
		}

		/* enum virDomainMemoryStatTags {
			VIR_DOMAIN_MEMORY_STAT_LAST           = VIR_DOMAIN_MEMORY_STAT_NR
			VIR_DOMAIN_MEMORY_STAT_SWAP_IN        = 0  //The total amount of data read from swap space (in kB).
			VIR_DOMAIN_MEMORY_STAT_SWAP_OUT       = 1  //The total amount of memory written out to swap space (in kB).
			VIR_DOMAIN_MEMORY_STAT_MAJOR_FAULT    = 2  //Page faults occur when a process makes a valid access to virtual memory that is not available. When servicing the page fault, if disk IO is required, it is considered a major fault. If not, it is a minor fault. These are expressed as the number of faults that have occurred.
			VIR_DOMAIN_MEMORY_STAT_MINOR_FAULT    = 3
			VIR_DOMAIN_MEMORY_STAT_UNUSED         = 4  //The amount of memory left completely unused by the system. Memory that is available but used for reclaimable caches should NOT be reported as free. This value is expressed in kB.
			VIR_DOMAIN_MEMORY_STAT_AVAILABLE      = 5  //The total amount of usable memory as seen by the domain. This value may be less than the amount of memory assigned to the domain if a balloon driver is in use or if the guest OS does not initialize all assigned pages. This value is expressed in kB.
			VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON = 6  //Current balloon value (in KB).
			VIR_DOMAIN_MEMORY_STAT_RSS            = 7  //Resident Set Size of the process running the domain. This value is in kB
			VIR_DOMAIN_MEMORY_STAT_USABLE         = 8  //How much the balloon can be inflated without pushing the guest system to swap, corresponds to 'Available' in /proc/meminfo
			VIR_DOMAIN_MEMORY_STAT_LAST_UPDATE    = 9  //Timestamp of the last update of statistics, in seconds.
			VIR_DOMAIN_MEMORY_STAT_NR             = 10 //The number of statistics supported by this version of the interface. To add new statistics, add them to the enum and increase this value.
		} */
		m := map[int32]string{
			int32(lv.DOMAIN_MEMORY_STAT_SWAP_IN):        "mem_swap_in",
			int32(lv.DOMAIN_MEMORY_STAT_SWAP_OUT):       "mem_swap_out",
			int32(lv.DOMAIN_MEMORY_STAT_MAJOR_FAULT):    "mem_major_fault",
			int32(lv.DOMAIN_MEMORY_STAT_MINOR_FAULT):    "mem_minor_fault",
			int32(lv.DOMAIN_MEMORY_STAT_UNUSED):         "mem_unused",
			int32(lv.DOMAIN_MEMORY_STAT_AVAILABLE):      "mem_available",
			int32(lv.DOMAIN_MEMORY_STAT_ACTUAL_BALLOON): "mem_actual_balloon",
			int32(lv.DOMAIN_MEMORY_STAT_RSS):            "mem_rss",
			int32(lv.DOMAIN_MEMORY_STAT_USABLE):         "mem_usable",
		}

		for _, stat := range stats {
			if val, ok := m[stat.Tag]; ok {
				fields[val] = stat.Val
			}
		}
		acc.AddFields("vm.data", fields, tags)

		GatherInterfaces(*connection, domain, acc, tags)

		GatherDisks(*connection, domain, acc, tags)

		domain.Free()
	}

	return nil
}

func GatherInterfaces(c lv.Connect, d lv.Domain, acc telegraf.Accumulator, tags map[string]string) error {
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
		tags["interface"] = iface.Name
		fields := map[string]interface{}{
			"rx_bytes":   iface.RxBytes,
			"rx_packets": iface.RxPkts,
			"rx_errs":    iface.RxErrs,
			"rx_drop":    iface.RxDrop,
			"tx_bytes":   iface.TxBytes,
			"tx_packets": iface.TxPkts,
			"tx_errs":    iface.TxErrs,
			"tx_drop":    iface.TxDrop,
		}
		acc.AddFields("vm.data", fields, tags)
	}
	delete(tags, "interface")
	return nil
}

func GatherDisks(c lv.Connect, d lv.Domain, acc telegraf.Accumulator, tags map[string]string) error {
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
		tags["disk"] = disk.Name
		fields := map[string]interface{}{
			"rd_req":   disk.RdReqs,
			"rd_bytes": disk.RdBytes,
			"wr_req":   disk.WrReqs,
			"wr_bytes": disk.WrBytes,
			"errs":     disk.Errors,
		}
		acc.AddFields("vm.data", fields, tags)
	}
	delete(tags, "disk")
	return nil
}

func init() {
	inputs.Add("libvirt", func() telegraf.Input {
		return &Libvirt{}
	})
}
