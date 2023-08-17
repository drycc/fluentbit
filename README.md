## Description
[![Build Status](https://woodpecker.drycc.cc/api/badges/drycc/fluentbit/status.svg)](https://woodpecker.drycc.cc/drycc/fluentbit)

Drycc (pronounced DAY-iss) is an open source PaaS that makes it easy to deploy and manage
applications on your own servers. Drycc builds on [Kubernetes](http://kubernetes.io/) to provide
a lightweight, [Heroku-inspired](http://heroku.com) workflow.

## About
This is an centos7 based image for running [fluentbit](http://fluentbit.io). It is built for the purpose of running on a kubernetes cluster.

This image is in with [workflow](https://github.com/drycc/workflow) v2 to send all log data to the [logger](https://github.com/drycc/logger) component.


## Chart values

```sh
helm show values oci://registry.drycc.cc/charts-testing/fluentbit
```

## Plugins

Fluent Bit currently supports integration of Golang plugins built as shared objects for output plugins only. The interface for the Golang plugins is currently under development but is functional.

### Drycc Output
Drycc output is a custom fluentbit plugin that was written to forward data directly to drycc components while filtering out data that we did not care about. We have 2 pieces of information we care about currently.

Logs from applications that are written to stdout within the container and the controller logs that represent actions against those applications. These logs are sent to an internal messaging system ([REIDS](https://redis.io/topics/streams-intro)) on a configurable topic. The logger component then reads those messages and stores the data in an ring buffer.
