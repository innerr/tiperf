# tiperf
Auto analyze prometheus metrics (and more) in TiDB, to help perf tuning

## Usage
Analyze all
```
tiperf timeline all
```
Output (just for now, rapidly revolving)
```
[2020-04-19 04:57:03 => 2020-04-20 05:19:03]
    ** started by workload changed, similarity 0.16
    2020-04-20 02:05:03 [tikv] -> down 17.1.4.135
    2020-04-20 05:18:03 [tikv] -> up 172.1.4.148
    ** lasted 24h22m0s, ended by workload changed, similarity 0.22
[2020-04-20 05:19:03 => 2020-04-20 05:21:03]
    ** started by workload changed, similarity 0.22
    ** lasted 2m0s, ended by workload changed, similarity 0.16
[2020-04-20 05:21:03 => 2020-04-20 16:11:03]
    ** started by workload changed, similarity 0.16
    ** lasted 10h50m0s
```

Analyze jitter and pike
```
tiperf timeline jitter pike
```

Analyze all but jitter
```
tiperf timeline all ~jitter
```

Analyze all, in specified prometheus address
```
tiperf --host 11.22.33.44 --port 5566 timeline all ~jitter
```

Get help
```
tiperf timeline
```

## How
First, `tiperf` split the timeline into many periods by analyzing workload changes,
then run detecting functions (the function/feature names passed from command line by user) to create infomation of each period.

A detecting function is like
```
Detector(data sources.Sources, period base.Period, found FoundEvents, con base.Console) (Events, error)
```
* `data` data query client, include prometheus or other clients
* `period` the start and end time to be analyzed
* `found` the events other detecting functions collected, dependencies can be configured when registering this function
* `con` stdin/stdout
* `Events` return the events of this function collected

The whole process is simple, get involed if you are interested
