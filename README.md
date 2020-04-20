# tiperf
Auto analyze prometheus metrics (and more) in TiDB, to help perf tuning

## Usage
Get help:
```
tiperf timeline
```
Analyze jitter and pike
```
tiperf timeline jitter pike
```
Analyze all features
```
tiperf timeline all
```
Analyze all but jitter
```
tiperf timeline all ~jitter
```
Analyze all, in specified prometheus address
```
tiperf --host 11.22.33.44 --port 5566 timeline all ~jitter
```

## How
First, tiperf split the timeline into many periods by analyzing workload changes,
then run detecting functions (the function/feature names passed from command line by user) to create infomation of each period.

A detecting function is like
```
Detector (sources sources.Sources, period base.Period, found FoundEvents, con base.Console) (Events, error)
```
* `sources` data query client, include prometheus or other ones
* `period` the start and end time to be analyzed
* `found` the events other detecting functions found, each function can specified it's dependencies when registering the function
* `con` stdin/stdout
* `Events` return the events of this function collected

The whole process is simple, welcome to join us!
