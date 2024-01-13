terraform {
  required_providers {
    k8scrd = {
      source = "dihedron.org/terraform/k8scrd"
    }
  }
}


provider "k8scrd" {
  host = "https://myendpoint.example.com/api"
  token = "12345567890ABCDEF"
  # example configuration here
}

resource "k8scrd_instance" "crontab_crd" {
  attributes = {
    name = "crontabs.stable.example.com"
    version = "v1"
    served = "true"
    storage = "true"
    scope = "Namespaced"
  }
  template = <<-EOT
    apiVersion: apiextensions.k8s.io/v1
    kind: CustomResourceDefinition
    metadata:
      # name must match the spec fields below, and be in the form: <plural>.<group>
      name: {{ .name }}
    spec:
      # group name to use for REST API: /apis/<group>/<version>
      group: stable.example.com
      # list of versions supported by this CustomResourceDefinition
      versions:
        - name: {{ .version }}
          # Each version can be enabled/disabled by Served flag.
          served: {{ .served }}
          # One and only one version must be marked as the storage version.
          storage: {{ .storage }}
          schema:
            openAPIV3Schema:
              type: object
              properties:
                spec:
                  type: object
                  properties:
                    cronSpec:
                      type: string
                    image:
                      type: string
                    replicas:
                      type: integer
      # either Namespaced or Cluster
      scope: {{ .scope }}
      names:
        # plural name to be used in the URL: /apis/<group>/<version>/<plural>
        plural: crontabs
        # singular name to be used as an alias on the CLI and for display
        singular: crontab
        # kind is normally the CamelCased singular type. Your resource manifests use this.
        kind: CronTab
        # shortNames allow shorter string to match your resource on the CLI
        shortNames:
        - ct  
    EOT
}

resource "k8scrd_instance" "crontab_object_1" {
  attributes = {
    kind = "crontabs.stable.example.com"
    name = "my-new-cron-object-1"
    cron = "* * * * */5"
    image = "my-awesome-cron-image"
  }
  template = <<-EOT
    apiVersion: "{{- .kind -}}"
    kind: CronTab
    metadata:
      name: {{ .name }}
    spec:
      cronSpec: "{{- .cron -}}"
      image: {{ .image }}
  EOT
}

