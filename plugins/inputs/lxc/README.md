# LXC Plugin

This plugin collects different metrics from LXC Containers. It collects three
different categories of metrics: CPU usage, IO performance and memory usage.
[This docker
blog](https://blog.docker.com/2013/10/gathering-lxc-docker-containers-metrics/)
is the basis for this plugin.

### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
# Description
[[inputs.example]]
  example_option = "example_value"
```

### Metrics:

Here you should add an optional description and links to where the user can
get more information about the measurements.

If the output is determined dynamically based on the input source, or there
are more metrics than can reasonably be listed, describe how the input is
mapped to the output.

- lxc-container
  - tags:
    - category (mem, cpu or io)
    - name (name of the container)
  - fields:
    - total\_inactive\_file
    - total\_inactive\_anon
    - unevictable
    - total\_active\_file
    - active\_anon
    - swap
    - total\_pgpgin
    - mapped\_file
    - total\_shmem
    - pgpgin
    - total\_active\_anon
    - cache
    - total\_pgpgout
    - dirty
    - total\_mapped\_file
    - shmem
    - total\_swap
    - inactive\_anon
    - active\_file
    - total\_writeback
    - rss
    - total\_unevictable
    - pgpgout
    - pgfault
    - total\_dirty
    - pgmajfault
    - total\_pgfault
    - total\_pgmajfault
    - inactive\_file
    - hierarchical\_memory\_limit
    - total\_rss\_huge
    - total\_rss
    - total\_cache
    - writeback
    - rss\_huge

    - user
    - total

    - blkio.sectors\_recursive
    - blkio.io\_service\_bytesTotal
    - blkio.io\_service\_bytes\_recursiveTotal
    - blkio.io\_servicedTotal
    - blkio.io\_serviced\_recursiveTotal
    - blkio.io\_queuedTotal
    - blkio.io\_queued\_recursiveTotal

### Sample Queries:

This section should contain some useful InfluxDB queries that can be used to
get started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min for the measurement in the last hour:
```
SELECT max(field1), mean(field1), min(field1) FROM measurement1 WHERE tag1=bar AND time > now() - 1h GROUP BY tag
```

### Example Output:

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter <plugin-name> --test` or use the `file` output to get
this information.

```
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```

