# Tyk Terraform Provider 
**This is a POC to test a tyk terraform provider.**

Run the following command to build and install the provider

```shell
go build -o terraform-provider-tyk
```

```shell
make install
```
## Test sample configuration

After building and installing the provider you can test it with the samples in the example folder.

```shell
cd examples && terraform init && terraform apply
```