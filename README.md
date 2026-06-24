## Rebrand of `tfctl` to `tfquery`.

`tfquery` was originally released as `tfctl` in this organization and repo.

In June 2026, HashiCorp [released](https://www.hashicorp.com/en/blog/introducing-tfctl-the-cli-for-hcp-terraform-and-tfe) its own CLI tool, [`tfctl-cli`](https://github.com/hashicorp/tfctl-cli).

To avoid confusion between the two projects, I have rebranded this project as **`tfquery`** (https://github.com/tfquery/tfquery).

While `tfquery` and HashiCorp's `tfctl-cli` have some overlapping capabilities—particularly around querying HCP Terraform and Terraform Enterprise resources—they are designed with different goals in mind. `tfctl-cli` focuses on managing HashiCorp platforms and services. `tfquery` focuses on querying, reporting, and exploring Terraform-related resources and metadata. Users may find value in using both tools together, depending on their workflow.
