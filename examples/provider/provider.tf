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


resource "k8scrd_instance" "example1" {
  attributes = {
    token = "1234-5678-90ABCDEF-12345678-90AB-CDEF"
    username = "myUsername1"
    password = "myS3cr3tP4$$w0rd!!"
    content = "Some other important stuff here for resource 1"
  }
  template = <<-EOT
  For running this template I'm going to use a security token (value: {{ .token }}")
  and if the token has not been provided, I could use a username and password combination
  for basic authentication (values: for the username it's {{ .username }} and for the 
  password that would be {{ .password }}).
  Last the content to store into the YAML is something like {{ .content }}.
  EOT
}


resource "k8scrd_instance" "example2" {
  attributes = {
    token = "1234-5678-90ABCDEF-12345678-90AB-CDEF"
    username = "myUsername2"
    password = "myS3cr3tP4$$w0rd!!"
    content = "Some other important stuff here for resource 2"
  }
  template = <<-EOT
  For running this template I'm going to use a security token (value: {{ .token }}")
  and if the token has not been provided, I could use a username and password combination
  for basic authentication (values: for the username it's {{ .username }} and for the 
  password that would be {{ .password }}).
  Last the content to store into the YAML is something like {{ .content }}.
  EOT
}

