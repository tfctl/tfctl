# tfctl-ps

> Show a summary of the given plan.
> Also available as: `tfctl summarize`

- Show only a summary of a Terraform plan.

`terraform plan | tfctl ps`

- Show the full plan output while also including a summary.

`terraform plan | tee >(tfctl ps)`
