# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or ar testable even if some parts are not relevant for the documentation.

* **provider/provider.tf** example file for the provider index page
* **data-sources/`full data source name`/data-source.tf** example file for the named data source page
* **resources/`full resource name`/resource.tf** example file for the named data source page

# How to build and run

Set the local .terraformc file for a local override:

```bash
$> cat ~/.terraformrc

provider_installation {

  dev_overrides {
      "dihedron.org/terraform/k8scrd" = "/data/workspaces/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

In order to build and install the provider into into `$GOBIN`, in the project root run

```bash
$> go install .
```

Then move into the `examples/provider` directory and run the following:

```bash
$> terraform apply
```

You can play with the `provider.tf` to create and destroy resources.
The provider does not invoke the real `kubectl`; it uses a custom application that simply dumps its command line and stdin to output, in directory `examples/kubectl`.