# Changelog v1.32

## Know before update


 - Add alerts with the recommended course of action to monitor LINSTOR, Piraeus-operator, capacity of storage-pools and resources states
 - Added Grafana dashboard to monitor LISNTOR cluster and DRBD resources
 - Multimaster clusters will automatically turn LINSTOR into HA-mode
 - OpenVPN will be migrated from using PVC to store certificates to Kubernetes secrets. PVC will still remain in the cluster as a backup. If you don't need it, you should manually delete it from the cluster.

## Features


 - **[ceph-csi]** Added new module ceph-csi [#426](https://github.com/deckhouse/deckhouse/pull/426)
    CephCSI allows dynamically provisioning Ceph volumes (RBD and CephFS) and attaching them to workloads.
 - **[linstor]** Grafana dashboard for LINSTOR [#1035](https://github.com/deckhouse/deckhouse/pull/1035)
    Added Grafana dashboard to monitor LISNTOR cluster and DRBD resources
 - **[linstor]** Alerts for LINSTOR [#1035](https://github.com/deckhouse/deckhouse/pull/1035)
    Add alerts with the recommended course of action to monitor LINSTOR, Piraeus-operator, capacity of storage-pools and resources states

## Fixes


 - **[cloud-provider-aws]** The necessary IAM policies for creating a peering connection have been added to the documentation. [#504](https://github.com/deckhouse/deckhouse/pull/504)
 - **[linstor]** LINSTOR module now supports high-availability [#1147](https://github.com/deckhouse/deckhouse/pull/1147)
    Multimaster clusters will automatically turn LINSTOR into HA-mode
 - **[node-local-dns]** Reworked health checking logic [#388](https://github.com/deckhouse/deckhouse/pull/388)
    Now Pods shouldn't crash unexpectedly now due to poor implementation of locking/probing.
 - **[openvpn]** Web interface changed to https://github.com/flant/ovpn-admin. Persistent storage has been replaced with Kubernetes secrets. Added HostPort inlet. [#522](https://github.com/deckhouse/deckhouse/pull/522)
    OpenVPN will be migrated from using PVC to store certificates to Kubernetes secrets. PVC will still remain in the cluster as a backup. If you don't need it, you should manually delete it from the cluster.
 - **[prometheus]** Set Grafana sample limit to 5000 [#1215](https://github.com/deckhouse/deckhouse/pull/1215)
 - **[upmeter]** Upmeter no longer exposes DNS queries to the Internet [#1256](https://github.com/deckhouse/deckhouse/pull/1256)
 - **[upmeter]** Fixed the calculation of groups uptime [#1144](https://github.com/deckhouse/deckhouse/pull/1144)

## Chore


 - **[docs]** Getting started - installing Deckhouse in a private environment. [#996](https://github.com/deckhouse/deckhouse/pull/996)
 - **[upmeter]** Remove redundant smoke-mini ingress [#1237](https://github.com/deckhouse/deckhouse/pull/1237)
 - **[upmeter]** Add User-Agent header to all requests [#1213](https://github.com/deckhouse/deckhouse/pull/1213)

