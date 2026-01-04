data "aws_ami" "amazon_linux_2023_arm64" {
  most_recent = true

  filter {
    name   = "name"
    values = ["al2023-ami-*"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  owners = ["amazon"]
}

resource "aws_security_group" "group" {
  vpc_id = local.vpc_id
  egress {
    cidr_blocks = ["0.0.0.0/0"]
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
  }
}

data "aws_iam_policy_document" "bastion_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["ec2.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "bastion" {
  assume_role_policy = data.aws_iam_policy_document.bastion_assume_role.json
}

resource "aws_iam_role_policy_attachment" "attachment" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.bastion.id
}

resource "aws_iam_instance_profile" "profile" {
  role = aws_iam_role.bastion.id
}

resource "aws_instance" "bastion" {
  ami                    = data.aws_ami.amazon_linux_2023_arm64.id
  instance_type          = "t4g.nano"
  iam_instance_profile   = aws_iam_instance_profile.profile.id
  subnet_id              = local.public_subnet_ids[0]
  vpc_security_group_ids = [aws_security_group.group.id]
  user_data              = <<-EOT
  sudo systemctl enable amazon-ssm-agent
  sudo systemctl start amazon-ssm-agent
  sudo systemctl status amazon-ssm-agent
  EOT
}