package netflow

import (
    "fmt"
    "strings"
    "net"
    "bufio"
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
    "errors"
    "strconv"
    "github.com/gophercloud/gophercloud"
    "github.com/gophercloud/gophercloud/openstack"
    "github.com/gophercloud/gophercloud/pagination"
    "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

type Netflow struct {
    IdentityEndpoint   string
    Username        string
    Password        string
    TenantName string
}

var sampleConfig = `
  IdentityEndpoint = ""
  username = "myuser"
  password = "mypassword"
  tenantname = "myuser"
`

func (n *Netflow) Description() string {
    return "Plugin to collect netflow data with pmacct and aggregate some metrics"
}

func (n *Netflow) SampleConfig() string {
    return sampleConfig
}
func (n *Netflow) GetIPUuidMappings() map[string]string {
    opts := gophercloud.AuthOptions{
        IdentityEndpoint: n.IdentityEndpoint,
        Username: n.Username,
        Password: n,Password,
        TenantName: n.TenantName,
    }
    provider, err := openstack.AuthenticatedClient(opts)
    if err != nil {
        return err
    }
    otheropts := gophercloud.EndpointOpts{Region: n.Region}
    client, err := openstack.NewComputeV2(provider, otheropts)
    if err != nil {
        return err
    }
    opts2 := servers.ListOpts{AllTenants: true}
    pager := servers.List(client, opts2)
    uuidIPMap := make(map[string]string)
    pager.EachPage(func(page pagination.Page) (bool, error) {
        serverList, err := servers.ExtractServers(page)

        fmt.Printf("size: %d\n", len(serverList)) 
        if err != nil {
            log.Fatal(err)
            return false, nil
        }
        for _, s := range serverList {

            fmt.Println(s.ID)
            for _, v:= range s.Addresses {
                for _,n := range v.([]interface{}) {
                    add, _ := n.(map[string]interface{})
                    if add["OS-EXT-IPS:type"] == "floating" {
                        uuidIPMap[s.ID] = add["addr"].(string)
                    }
                }
            }
        }
        return true, nil
    })
    return uuidIPMap
}

func (n *Netflow) Gather(acc telegraf.Accumulator) error {
//    users := t.GetUserInfo()
//    for _,u := range users {
//        if c, _ := strconv.Atoi(u["client_type"]); c == 0 {
//            tags := map[string]string{
//                "unique_identifier" : u["client_unique_identifier"],
//                "input_muted" : u["client_input_muted"],
//                "away" : u["client_away"],
//                "output_muted" : u["client_output_muted"],
//            }
//            values := map[string]interface{}{
//                "nickname" : u["client_nickname"],
//                "packets_received_total" : u["connection_packets_received_total"],
//                "packets_sent_total" : u["connection_packets_sent_total"],
//                "ip" : u["connection_client_ip"],
//                "lastconnected" : u["client_lastconnected"],
//            }
//            acc.AddFields("user", values, tags)
//        }
//    }
    return nil
}

func init() {
    inputs.Add("netflow", func() telegraf.Input { return &Netflow{} })
}
