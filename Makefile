build:
	gb build -f -F

stage:
	aws ec2 run-instances --image-id ami-b5a7ea85 --user-data "This is some user data" --count 2 --key-name greg --instance-type t2.micro
