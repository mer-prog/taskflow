variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "project" {
  description = "Project name used for resource naming"
  type        = string
  default     = "taskflow"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "db_password" {
  description = "Password for the RDS PostgreSQL instance"
  type        = string
  sensitive   = true
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "container_image" {
  description = "Docker image URI for the API container"
  type        = string
}

variable "container_cpu" {
  description = "Fargate task CPU units (256 = 0.25 vCPU)"
  type        = number
  default     = 256
}

variable "container_memory" {
  description = "Fargate task memory in MiB"
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "Number of ECS tasks to run"
  type        = number
  default     = 2
}

variable "jwt_secret" {
  description = "JWT signing secret"
  type        = string
  sensitive   = true
}

variable "cors_origin" {
  description = "Allowed CORS origin(s), comma-separated"
  type        = string
  default     = "https://taskflow-app.vercel.app"
}
