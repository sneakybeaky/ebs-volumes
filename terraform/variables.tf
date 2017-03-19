variable "public_key_path" {
  description = <<DESCRIPTION
Path to the SSH public key to be used for authentication.
Ensure this keypair is added to your local SSH agent so provisioners can
connect.

Example: ~/.ssh/terraform.pub
DESCRIPTION
}

variable "key_name" {
  description = "Desired name of AWS key pair"
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default = "eu-west-2"
}

variable "aws_availability_zone" {
  description = "AWS availability zone to use."
  default = "eu-west-2a"
}

# Ubuntu 16.04 LTS (x64) hvm-ssd release 20170307
variable "aws_amis" {
  type = "map"
  default = {
    eu-west-1 = "ami-971238f1"
    eu-west-2 = "ami-ed908589"
    us-east-1 = "ami-2757f631"
    us-west-1 = "ami-44613824"
    us-west-2 = "ami-ed908589"
  }
}
