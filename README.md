# Tyk Terraform Provider 

**This is a POC to test a tyk terraform provider.**

## OpenTofu
Following recent [changes in Terraform's MPL](https://opentofu.org/manifesto), we've begun exploring [OpenTofu](https://opentofu.org/), a [fork](https://github.com/opentofu/opentofu) of TF and true open-source managed by the Linux Foundation. 
You can watch this brief [video](https://www.loom.com/share/526bff9e6f0b48b8b0d9bdf477cd8ec8) to learn and see that OpenTofu binary works seamlessly with our Tyk TF provider to provision Tyk API Management resources on Tyk Cloud.

## Get started

### Build Tyk TF Provider
Run the following command to build and install the provider

```shell
go build -o terraform-provider-tyk
```

```shell
make install
```

### Test sample configuration
 
You will need your tyk cloud email and Tyk cloud password.
You can save the email in your environment variable with the name: `TYK_EMAIL`
The password can also be saved in your environment variable with the name: `TYK_PASSWORD`

Sample:
```shell
  export TYK_EMAIL="email here"
  export TYK_PASSWORD="password here"
```

After building and installing the provider you can test it with the samples in the example folder.

```shell
cd examples && terraform init && terraform apply
```

## Project Structure
- [Examples](./examples/)
     - Contains the example TF file you can use to build a complete and new deployment in Tyk Cloud, as explained in this [README](#test-sample-configuration)
- [tyk](./tyk)
    - Tyk everyday users, who simply wish to use TF with [Tyk cloud](https://tyk.io/cloud/) services, can safely bypass this directory. 
    - You should delve into it exclusively if you aspire to contribute to this open-source project.
     The TF SDK and Tyk SDK serve as the bridge to connect to Tyk cloud infrastructure, creating resources, modifying or deleting.
     
