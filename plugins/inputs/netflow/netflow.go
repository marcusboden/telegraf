package main

import (
  "fmt"
  "log"
  "github.com/gophercloud/gophercloud"
  "github.com/gophercloud/gophercloud/openstack"
  "github.com/gophercloud/gophercloud/pagination"
  "github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
)

func main() {
  opts := gophercloud.AuthOptions{
    IdentityEndpoint: "https://cloud.gwdg.de:5000/",
    Username: "admin",
    Password: "QjtTCfTj87Gq",
    TenantName: "admin",
  }
  provider, err := openstack.AuthenticatedClient(opts)
  if err != nil {
    log.Fatal(err)
    return
  }
  otheropts := gophercloud.EndpointOpts{Region: "RegionOne"}
  client, err := openstack.NewComputeV2(provider, otheropts)
  if err != nil {
    log.Fatal(err)
    return
  }
  opts2 := servers.ListOpts{AllTenants: true}
  pager := servers.List(client, opts2)
  pager.EachPage(func(page pagination.Page) (bool, error) {
    serverList, err := servers.ExtractServers(page)

    fmt.Printf("size: %d\n", len(serverList)) 
    if err != nil {
      log.Fatal(err)
      return false, nil
    }
    for _, s := range serverList {

      fmt.Println(s.ID, s.Name, s.Status)

    }
    return true, nil
  })
}
