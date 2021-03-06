# Specify the provider and access details
provider "aws" {
  region = "${var.aws_region}"
}

data "aws_availability_zones" "available" {}

# Retrieve the AZ where we want to create network resources
# This must be in the region selected on the AWS provider.
data "aws_availability_zone" "primary" {
  name = "${data.aws_availability_zones.available.names[0]}"
}

# Create a VPC to launch our instances into
resource "aws_vpc" "default" {
  cidr_block = "${cidrsubnet("10.0.0.0/12", 4, var.region_number[data.aws_availability_zone.primary.region])}"
}

# Create an internet gateway to give our subnet access to the outside world
resource "aws_internet_gateway" "default" {
  vpc_id = "${aws_vpc.default.id}"
}

# Grant the VPC internet access on its main route table
resource "aws_route" "internet_access" {
  route_table_id = "${aws_vpc.default.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.default.id}"
}

# Create a subnet to launch our instances into
resource "aws_subnet" "default" {
  vpc_id = "${aws_vpc.default.id}"
  cidr_block = "${cidrsubnet(aws_vpc.default.cidr_block, 4, var.az_number[data.aws_availability_zone.primary.name_suffix])}"
  map_public_ip_on_launch = true
  availability_zone = "${data.aws_availability_zone.primary.name}"
}


# Our default security group to access
# the instances over SSH and HTTP
resource "aws_security_group" "default" {
  name = "terraform_example"
  description = "Used in the terraform"
  vpc_id = "${aws_vpc.default.id}"

  # SSH access from anywhere
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  # outbound internet access
  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = [
      "0.0.0.0/0"]
  }
}


resource "aws_key_pair" "auth" {
  key_name = "${var.key_name}"
  public_key = "${file(var.public_key_path)}"
}

# Ubuntu 16.04 LTS (x64) hvm-ssd
data "aws_ami" "server_ami" {
  most_recent      = true

  filter {
    name = "architecture"
    values = ["x86_64"]
  }

  filter {
    name = "block-device-mapping.volume-type"
    values = ["gp2"]
  }

  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "name"
    values = ["*ubuntu-xenial-16.04-amd64-server-*"]
  }

  owners     = ["099720109477"]
}

resource "aws_instance" "web" {

  availability_zone = "${data.aws_availability_zone.primary.name}"

  # The connection block tells our provisioner how to
  # communicate with the resource (instance)
  connection {
    # The default username for our AMI
    user = "ubuntu"

    # The connection will use the local SSH agent for authentication.
  }

  iam_instance_profile = "${aws_iam_instance_profile.test_profile.id}"

  instance_type = "m4.large"

  # Lookup the correct AMI based on the region
  # we specified
  #ami = "${lookup(var.aws_amis, var.aws_region)}"
  ami = "${data.aws_ami.server_ami.id}"

  # The name of our SSH keypair we created above.
  key_name = "${aws_key_pair.auth.id}"

  # Our Security group to allow HTTP and SSH access
  vpc_security_group_ids = [
    "${aws_security_group.default.id}"]

  # We're going to launch into the same subnet as our ELB. In a production
  # environment it's more common to have a separate private subnet for
  # backend instances.
  subnet_id = "${aws_subnet.default.id}"

  tags {
    Name = "test",
    foo = "bar"
    "volume_/dev/sdf" = "${aws_ebs_volume.volume1.id}"
    "volume_/dev/sdg" = "${aws_ebs_volume.volume2.id}"
    "detach_volumes" = "true"
  }

}

resource "aws_ebs_volume" "volume1" {
  size = 1
  availability_zone = "${data.aws_availability_zone.primary.name}"
}

resource "aws_ebs_volume" "volume2" {
  size = 1
  availability_zone = "${data.aws_availability_zone.primary.name}"
}

resource "aws_iam_instance_profile" "test_profile" {
  name = "testprofile"
  roles = [
    "${aws_iam_role.test_role.id}"]
}

resource "aws_iam_role_policy" "test_policy" {
  name = "test_policy"
  role = "${aws_iam_role.test_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:DescribeTags",
        "ec2:DescribeVolumes",
        "ec2:AttachVolume",
        "ec2:DetachVolume"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test_role" {
  name = "test_role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}