# 在线评测系统 (Online Judge System)
这是一个基于Go的在线评测系统，支持代码提交、评测、结果反馈和题解发布等功能。

## 项目概述
该系统实现了完整的代码评测流程，包括用户代码提交、并发运行、结果比对、超时检测和反馈展示。支持本地和远程评测模式，保障了系统的灵活性和可扩展性。

## 评测机实现
[go-judge](https://github.com/crazyfrankie/go-judge)

## 技术栈
- 后端框架：Gin + GORM
- 数据存储：MySQL + Redis + MongoDB
- 消息队列：Kafka
- 监控系统：Prometheus

## 核心功能
### 用户认证与权限管理
- 基于Cookie + JWT的双Token机制
- RBAC 权限控制（普通用户/管理员）
- 用户状态无感刷新

### 代码评测与资源隔离
- 基于namespace和cgroup实现资源隔离 
- 支持本地与远程评测 
- 并发执行用户提交的代码 
- 智能比对运行结果并反馈

### 系统优化
- Redis缓存热门数据与查询结果 
- Lua脚本优化并发操作 
- 滑动窗口限流策略 
- 基于雪花算法的唯一ID生成 
- MongoDB分片存储大文本数据 

### 可靠性保障
- 短信登录Failover机制
- Prometheus监控系统状态

## 快速开始
环境需求
- Go 1.23+
- Docker & Docker Compose
- MySQL & Redis & MongoDB