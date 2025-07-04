## Description
[![Build Status](https://woodpecker.drycc.cc/api/badges/drycc/fluentbit/status.svg)](https://woodpecker.drycc.cc/drycc/fluentbit)

Drycc (pronounced DAY-iss) is an open source PaaS that makes it easy to deploy and manage
applications on your own servers. Drycc builds on [Kubernetes](http://kubernetes.io/) to provide
a lightweight, [Heroku-inspired](http://heroku.com) workflow.

## About
This is an debian based image for running [fluentbit](http://fluentbit.io). It is built for the purpose of running on a kubernetes cluster.

This image is in with [workflow](https://github.com/drycc/workflow) v2 to send all log data to the [quickwit](https://github.com/drycc/quickwit) component.


## Helm Chart
Once you have installed the Helm client, you can deploy a Fluentbit Chart into a Kubernetes cluster.

### Chart repos

* Stable Charts: oci://registry.drycc.cc/charts/fluentbit
* Testing Charts: oci://registry.drycc.cc/charts-testing/fluentbit

### Chart values
```sh
helm show values oci://registry.drycc.cc/charts-testing/fluentbit
```
