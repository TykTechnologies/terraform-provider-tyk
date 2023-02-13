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
 
You will need your tyk cloud email and Tyk cloud password.

You can save the email in your environment variable with the name : TYK_EMAIL

The password can also be saved in your environment variable with the name : TYK_PASSWORD

Sample:
```shell
  export TYK_EMAIL="email here"
  export TYK_PASSWORD="password here"
```

After building and installing the provider you can test it with the samples in the example folder.

```shell
cd examples && terraform init && terraform apply
```