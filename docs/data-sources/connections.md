---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kinde_connections Data Source - kinde"
subcategory: ""
description: |-
  Use this data source to list available connections.
---

# kinde_connections (Data Source)

Use this data source to list available connections.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter` (String) Filter connections by type. Valid values are: `builtin`, `custom`, `all`. Defaults to `all`.

### Read-Only

- `connections` (Attributes List) (see [below for nested schema](#nestedatt--connections))

<a id="nestedatt--connections"></a>
### Nested Schema for `connections`

Read-Only:

- `display_name` (String)
- `id` (String)
- `name` (String)
- `strategy` (String)
