# 自动化Minikube集群管理

本E2E测试框架现在包含了自动化的Minikube集群管理功能，确保每个测试都在全新、隔离的环境中运行。

## 功能特性

### 🔄 自动集群管理
- **自动创建**：测试开始时自动创建全新的minikube集群
- **自动清理**：测试结束后自动删除整个集群
- **完全隔离**：每次测试都在独立的环境中运行

### 📋 前置条件检查
- Go环境验证
- Minikube安装检查
- Docker守护程序状态
- Kubectl可用性验证

### 🚀 一键测试执行
```bash
# 运行所有测试（包含自动minikube管理）
./run_tests.sh

# 运行特定测试类型
./run_tests.sh -t http -v

# 使用自定义超时时间
./run_tests.sh -T 30m
```

## 测试流程

### 完整的端到端流程
1. **环境检查** - 验证所有必需工具
2. **集群创建** - 启动全新minikube集群（约2-3分钟）
3. **证书生成** - 创建测试所需的TLS证书
4. **Secret创建** - 在Kubernetes中创建证书Secret
5. **Symphony部署** - 使用证书配置启动Symphony
6. **测试执行** - 运行remote agent通信测试
7. **自动清理** - 删除minikube集群和所有资源

### 时间预期
- **HTTP测试**: ~10-15分钟
- **MQTT测试**: ~15-20分钟
- **完整测试套件**: ~25-30分钟

## 关键改进

### 🎯 解决的问题
- ✅ **证书信任问题**: 自动创建正确的Certificate Secret
- ✅ **环境一致性**: 每次测试都使用相同的集群配置
- ✅ **依赖管理**: 自动处理minikube和Symphony的启动
- ✅ **资源冲突**: 完全隔离的测试环境

### 🔧 自动化配置
```yaml
# Minikube集群配置
Memory: 4096MB
CPUs: 2
Disk: 20GB
Driver: docker
Wait: all components ready

# Symphony Helm配置
remoteAgent.remoteCert.used: true
remoteAgent.remoteCert.trustCAs.secretName: client-cert-secret
remoteAgent.remoteCert.trustCAs.secretKey: client-cert-key  
remoteAgent.remoteCert.subjects: MyRootCA
http.enabled: true
```

### 📦 Certificate管理
```bash
# CA Secret (cert-manager namespace)
Name: client-cert-secret
Key: client-cert-key -> CA证书内容

# Client Secret (测试namespace)  
Name: remote-agent-client-secret
Keys: client.crt, client.key

# 客户端证书Subject
CN=MyRootCA (固定值，匹配Symphony信任配置)
```

## 使用指南

### 基本用法
```bash
cd test/e2e/remote-agent-integration

# 运行HTTP通信测试
./run_tests.sh -t http

# 运行MQTT通信测试
./run_tests.sh -t mqtt

# 运行所有测试（推荐）
./run_tests.sh
```

### 高级选项
```bash
# 启用详细输出
./run_tests.sh -v

# 自定义超时时间（适用于慢机器）
./run_tests.sh -T 35m

# 仅运行HTTP测试且详细输出
./run_tests.sh -t http -v -T 20m
```

### 故障排除
```bash
# 如果测试失败，检查前置条件
./run_tests.sh --help

# 手动验证环境
minikube version
docker info
kubectl version --client
```

## 环境要求

### 必需工具
- **Go** 1.19+
- **Minikube** 1.25+
- **Docker** 20.10+
- **Kubectl** 1.24+

### 系统资源
- **内存**: 至少6GB可用（4GB给minikube + 2GB给系统）
- **CPU**: 至少4核心（2核给minikube + 2核给系统）
- **磁盘**: 至少30GB可用空间

### 网络要求
- Docker守护程序运行
- 能够拉取Docker镜像
- 无防火墙阻止minikube通信

## 错误处理

### 常见问题

1. **Minikube启动失败**
   ```bash
   # 清理并重试
   minikube delete
   ./run_tests.sh
   ```

2. **Docker权限问题**
   ```bash
   # 确保用户在docker组中
   sudo usermod -aG docker $USER
   newgrp docker
   ```

3. **资源不足**
   ```bash
   # 释放内存并重试
   docker system prune -f
   ./run_tests.sh -T 40m
   ```

## 架构优势

### 🔒 完全隔离
每个测试运行在独立的minikube集群中，消除了测试间的相互影响。

### 🎯 真实环境
使用完整的Kubernetes集群和真实的Symphony部署，确保测试结果的可信度。

### ⚡ 并行安全
多个测试实例可以并行运行而不会相互干扰。

### 🔄 可重复性
每次测试都在相同的初始状态下开始，确保结果一致性。

---

这个自动化框架使remote agent的E2E测试变得简单、可靠且真正端到端。无需手动管理集群或担心环境状态，只需运行测试脚本即可！
