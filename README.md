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

To fetch the results of all the sacns from a scansettingbinding, simply do:

```
$ oc compliance fetch-raw scansettingbinding nist-moderate -o resultsdir/
```

it'll be a similar operator if you want to use `ComplianceSuite` or
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

Installing
----------

There is an `install` target that's already set up in the Makefile for this
project.

However, as any other `oc` plugin, you may just copy the binary to the same
directory where the `oc` binary is.
