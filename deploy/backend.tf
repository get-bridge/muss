terraform {
  backend "s3" {
    bucket = "bridge-tfstate"
    key    = "muss/terraform/terraform.tfstate"
    region = "us-west-2"
    acl    = "bucket-owner-full-control"

    role_arn = "arn:aws:iam::127178877223:role/xacct/ops-admin"
  }
}
