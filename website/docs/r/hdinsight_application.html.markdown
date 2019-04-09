---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_hdinsight_application"
sidebar_current: "docs-azurerm-resource-hdinsight-application"
description: |-
  Manages an Application within a HDInsight Cluster.
---

# azurerm_hdinsight_application

Manages an Application within a HDInsight Cluster.

## Example Usage

```hcl
data "azurerm_hdinsight_cluster" "example" {
  name                = "example-cluster"
  resource_group_name = "example-resources"
}

resource "azurerm_hdinsight_application" "example" {
  name                   = "custom-application"
  cluster_id             = "${data.azurerm_hdinsight_hadoop_cluster.example.id}"
  marketplace_identifier = "CustomApplication"
  vm_size                = "Standard_D4_V2"

  install_script_action {
    name  = "say-hello"
    uri   = "https://gist.githubusercontent.com/tombuildsstuff/74ff75620a83cf2a737843920185dbc2/raw/8217fbbcf9728e23807c19a35f65136351e6da7a/hello.sh"
    roles = [ "edgenode" ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) Specifies the ID of the HDInsight Cluster where this Application should be installed. Changing this forces a new resource to be created.

* `name` - (Required) Specifies the name of the HDInsight Application. Changing this forces a new resource to be created.

* `marketplace_identifier` - (Required) Specifies the Marketplace Identifier for this HDInsight Application. Changing this forces a new resource to be created

* `vm_size` - (Required) Specifies the size of the Virtual Machine used as the Edge Node. Changing this forces a new resource to be created.


---

* `https_endpoint` - (Optional) One or more `https_endpoint` blocks as defined below.

* `install_script_action` - (Optional) One or more `install_script_action` blocks as defined below.

* `uninstall_script_action` - (Optional) One or more `uninstall_script_action` blocks as defined below.

---

A `https_endpoint` block supports the following:

* `destination_port` - (Required) The destination port to connect to. Changing this forces a new resource to be created.

* `access_modes` - (Optional) A list of Access Modes for this Application. Changing this forces a new resource to be created.

* `public_port` - (Optional) The public port to connect to. Changing this forces a new resource to be created.

---

A `install_script_action` block supports the following:

* `name` - (Required) A name for this Script Action. Changing this forces a new resource to be created.

* `roles` - (Required) The HDInsight Cluster Roles where this script should run. Possible values are `edgenode`, `headnode`, `workernode` and `zookeepernode`.. Changing this forces a new resource to be created.

* `uri` - (Required) The URI to the script which should be run. Changing this forces a new resource to be created.

-> **NOTE:** The script available at this URI must be idempotent.

---

A `uninstall_script_action` block supports the following:

* `name` - (Required) A name for this Script Action. Changing this forces a new resource to be created.

* `roles` - (Required) The HDInsight Cluster Roles where this script should run. Possible values are `edgenode`, `headnode`, `workernode` and `zookeepernode`.. Changing this forces a new resource to be created.

* `uri` - (Required) The URI to the script which should be run. Changing this forces a new resource to be created.

-> **NOTE:** The script available at this URI must be idempotent.


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the HDInsight Application.

## Import

HDInsight Application's can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_hdinsight_hadoop_cluster.test /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.HDInsight/clusters/cluster1/applications/application1
```
