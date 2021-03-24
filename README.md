oc compliance
=============

This is an `oc` plugin that is meant to be used with the
[compliance-operator](https://github.com/openshift/compliance-operator).

It's a set of utilities that make it easier to use the operator.

Subcommands
-----------

### fetch-raw

Helps download the raw compliance results from the Persistent Volume that
the operator stores them at.

To fetch the results of all the scans from a scansettingbinding, simply do:

```
$ oc compliance fetch-raw scansettingbinding nist-moderate -o resultsdir/
```

It'll be a similar operator if you want to use `ComplianceSuite` or
`ComplianceScan` objects.

### rerun-now

Forces the scan or set of scans to re-run on command instead of waiting for
them to be sheduled.

```
$ oc compliance rerun-now scansettingbinding nist-moderate
```

### controls

Creates a report of what compliance standards and controls will a benchmark
fulfil.

```
$ oc compliance controls profile rhcos4-moderate
```

### bind

Creates a `ScanSettingBinding` or the given parameters

```
$ oc compliance bind -N my-binding profile/rhcos4-moderate
```

* `--dry-run` is also supported. This will print the yaml that's needed to create the object.

### view-result

Gathers information in one place about a specific compliance result.

```
oc compliance view-result rhcos4-e8-worker-sysctl-kernel-kptr-restrict
+----------------------+---------------------------------------------------------------------------------+
|         KEY          |                                      VALUE                                      |
+----------------------+---------------------------------------------------------------------------------+
| title                | Restrict Exposed Kernel                                                         |
|                      | Pointer Addresses Access                                                        |
+----------------------+---------------------------------------------------------------------------------+
| status               | PASS                                                                            |
+----------------------+---------------------------------------------------------------------------------+
| severity             | medium                                                                          |
+----------------------+---------------------------------------------------------------------------------+
| description          | <code>kernel.kptr_restrict</code><pre>$ sudo sysctl -w                          |
|                      | kernel.kptr_restrict=1</pre><code>/etc/sysctl.d</code><pre>kernel.kptr_restrict |
|                      | = 1</pre>:                                                                      |
+----------------------+---------------------------------------------------------------------------------+
| rationale            | <code>seq_printf()</code>)                                                      |
|                      | exposes&#xA;kernel writeable                                                    |
|                      | structures that can contain                                                     |
|                      | functions pointers. If a write                                                  |
|                      | vulnereability occurs&#xA;in                                                    |
|                      | the kernel allowing a                                                           |
|                      | write access to any of this                                                     |
|                      | structure, the kernel can be                                                    |
|                      | compromise. This&#xA;option                                                     |
|                      | disallow any program withtout                                                   |
|                      | the CAP_SYSLOG capability from                                                  |
|                      | getting the kernel pointers                                                     |
|                      | addresses,&#xA;replacing them                                                   |
|                      | with 0.                                                                         |
+----------------------+---------------------------------------------------------------------------------+
| NIST-800-53 Controls | SC-30, SC-30(2), SC-30(5),                                                      |
|                      | CM-6(a)                                                                         |
+----------------------+---------------------------------------------------------------------------------+
| Avalailable Fix      | Yes                                                                             |
+----------------------+---------------------------------------------------------------------------------+
| Fix Object           | ---                                                                             |
|                      |                                                                                 |
|                      | apiVersion:                                                                     |
|                      | machineconfiguration.openshift.io/v1                                            |
|                      |                                                                                 |
|                      | kind: MachineConfig                                                             |
|                      |                                                                                 |
|                      | spec:                                                                           |
|                      |                                                                                 |
|                      |   config:                                                                       |
|                      |                                                                                 |
|                      |     ignition:                                                                   |
|                      |                                                                                 |
|                      |       version: 3.1.0                                                            |
|                      |                                                                                 |
|                      |     storage:                                                                    |
|                      |                                                                                 |
|                      |       files:                                                                    |
|                      |                                                                                 |
|                      |       - contents:                                                               |
|                      |                                                                                 |
|                      |           source:                                                               |
|                      | data:,kernel.kptr_restrict%3D1                                                  |
|                      |                                                                                 |
|                      |         mode: 420                                                               |
|                      |                                                                                 |
|                      |         path:                                                                   |
|                      | /etc/sysctl.d/75-sysctl_kernel_kptr_restrict.conf                               |
|                      |                                                                                 |
|                      |                                                                                 |
+----------------------+---------------------------------------------------------------------------------+
| Result Object Name   | rhcos4-e8-worker-sysctl-kernel-kptr-restrict                                    |
+----------------------+---------------------------------------------------------------------------------+
| Rule Object Name     | rhcos4-sysctl-kernel-kptr-restrict                                              |
+----------------------+---------------------------------------------------------------------------------+
| Remediation Created  | No                                                                              |
+----------------------+---------------------------------------------------------------------------------+
```

### fetch-fixes

Helps download the remediations the Compliance Operator recommends. These are
stored as YAML files in the filesystem, so one would then be able to apply them to a
cluster.

Note that MachineConfigs are not complete and require extra parsing:

* Setting an appropriate name
* Setting metadata for the pool they'll apply to

```
oc compliance fetch-fixes profile ocp4-cis -o tmp/
No fixes to persist for rule 'ocp4-accounts-restrict-service-account-tokens'
...
No fixes to persist for rule 'ocp4-api-server-audit-log-maxbackup'
Persisted rule fix to tmp/ocp4-api-server-audit-log-maxsize.yaml
Persisted rule fix to tmp/ocp4-api-server-encryption-provider-cipher.yaml
Persisted rule fix to tmp/ocp4-api-server-encryption-provider-config.yaml
```

Installing
----------

There is an `install` target that's already set up in the Makefile for this
project.

However, as any other `oc` plugin, you may just copy the binary to the same
directory where the `oc` binary is.
