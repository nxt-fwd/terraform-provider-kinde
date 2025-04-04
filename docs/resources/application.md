---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kinde_application Resource - kinde"
subcategory: ""
description: |-
  Applications facilitates the interface for users to authenticate against. See documentation https://docs.kinde.com/build/applications/about-applications/ for more details.
---

# kinde_application (Resource)

Applications facilitates the interface for users to authenticate against. See [documentation](https://docs.kinde.com/build/applications/about-applications/) for more details.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the application. Currently, there is no way to change this via the management application.
- `type` (String) Type of the application

### Optional

- `homepage_uri` (String) The homepage URI of the application.
- `login_uri` (String) The login URI of the application.
- `logout_uris` (List of String) The logout URIs of the application.
- `redirect_uris` (List of String) The redirect URIs of the application.

### Read-Only

- `client_id` (String) Client id of the application
- `client_secret` (String, Sensitive) Client secret of the application
- `id` (String) ID of the application
