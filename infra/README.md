# Infrastructure

This directory contains AWS infrastructure-as-code for the **production-grade** deployment of TaskFlow.

> **Note:** The demo environment uses **Render** (backend + PostgreSQL) and **Vercel** (frontend).
> The Terraform code here represents the intended AWS production architecture and is not actively deployed.

## Architecture

See [architecture.md](./architecture.md) for a detailed diagram and explanation.

## Terraform

```bash
cd terraform
terraform init
terraform plan -var="db_password=YOUR_PASSWORD" -var="container_image=YOUR_ECR_IMAGE"
terraform apply
```

## Files

| File | Description |
|------|-------------|
| `terraform/main.tf` | VPC, ECS, RDS, ALB, Security Groups |
| `terraform/variables.tf` | Input variables |
| `terraform/outputs.tf` | Output values (ALB DNS, RDS endpoint) |
| `ecs-task-definition.json` | ECS Fargate task definition |
| `architecture.md` | Architecture diagram and cost estimate |
