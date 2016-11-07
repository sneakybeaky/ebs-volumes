output "private_ip" {
  value = "${aws_instance.web.private_ip}"
}

output "public_ip" {
  value = "${aws_instance.web.public_ip}"
}