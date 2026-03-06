# AWS Architecture

## Architecture Diagram

```
                         Internet
                            |
                     +------+------+
                     |   Route 53  |
                     +------+------+
                            |
                     +------+------+
                     |     ALB     |
                     | (port 443)  |
                     +------+------+
                            |
               +------------+------------+
               |                         |
        +------+------+          +------+------+
        | ECS Fargate |          | ECS Fargate |
        |  (Task 1)   |          |  (Task 2)   |
        |  Go API     |          |  Go API     |
        |  :8080      |          |  :8080      |
        +------+------+          +------+------+
               |                         |
               +------------+------------+
                            |
                     +------+------+
                     |   RDS       |
                     | PostgreSQL  |
                     | (db.t3.micro)|
                     +-------------+

    VPC: 10.0.0.0/16
    ├── Public Subnets  (10.0.1.0/24, 10.0.2.0/24)  → ALB
    └── Private Subnets (10.0.10.0/24, 10.0.11.0/24) → ECS, RDS
```

## AWS Services

| Service | Role | Details |
|---------|------|---------|
| **VPC** | Network isolation | 2 AZs, public + private subnets, NAT Gateway |
| **ALB** | Load balancer | HTTPS termination, health checks, WebSocket support |
| **ECS Fargate** | Container orchestration | Serverless, auto-scaling, no EC2 management |
| **RDS PostgreSQL** | Database | db.t3.micro, Multi-AZ optional, automated backups |
| **ECR** | Container registry | Docker image storage |
| **Route 53** | DNS | Domain routing (optional) |
| **ACM** | SSL/TLS | Certificate management for HTTPS |
| **Security Groups** | Firewall | ALB: 80/443 inbound; ECS: 8080 from ALB; RDS: 5432 from ECS |

## Cost Estimate (Monthly)

| Service | Specification | Estimated Cost |
|---------|--------------|---------------|
| ECS Fargate | 0.25 vCPU, 0.5 GB x 2 tasks | ~$18 |
| RDS | db.t3.micro, 20 GB gp3 | ~$15 |
| ALB | 1 LCU average | ~$22 |
| NAT Gateway | 1 AZ | ~$32 |
| ECR | < 1 GB | ~$0.10 |
| **Total** | | **~$87/month** |

> For cost optimization: use a single NAT Gateway, single-AZ RDS, and minimum Fargate sizing.

## Security

- ECS tasks run in private subnets (no public IP)
- RDS accessible only from ECS security group
- ALB handles TLS termination
- Environment variables managed via ECS task definition secrets
- Database credentials stored in AWS Secrets Manager (recommended)
