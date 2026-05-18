# Trae Docker Usage

本文档基于 `docs/如何将日常使用的仓库环境构建成dockerfile，并用Trae启动容器？.docx`
整理，提供一套适用于当前 `logsift` 项目的 Trae SSH 容器方案。

## 文件位置

- `environment/Dockerfile`
- `environment/ssh_plugin/entrypoint.sh`
- `environment/ssh_plugin/install_ssh.sh`

这套配置不改动仓库根目录现有 `Dockerfile` 的用途，只额外提供一个给 Trae 连接的开发容器。

## 为什么单独做一套环境

- Trae 通过 SSH 连接容器，需要容器内运行 `sshd`
- Trae 远端服务依赖 `glibc`，因此避免使用 Alpine / musl 基础镜像
- 当前方案使用 `golang:1.24-bookworm`，和项目本身的 Go 环境一致，也更适合 Trae

## 构建镜像

在仓库根目录执行：

```bash
docker build -f environment/Dockerfile -t logsift-trae .
```

如果你的网络环境需要代理，可以把代理通过构建参数传进去：

```bash
docker build \
  --build-arg http_proxy=http://host:port \
  --build-arg https_proxy=http://host:port \
  -f environment/Dockerfile \
  -t logsift-trae .
```

## 启动容器

```bash
docker run -d \
  --name logsift-trae \
  -p 2222:22 \
  -e SSH_PASSWORD=password \
  logsift-trae
```

默认账号信息：

- 用户名：`root`
- 密码：`password`
- SSH 端口：`2222`

如果想换密码或端口，可以在运行时覆盖：

```bash
docker run -d \
  --name logsift-trae \
  -p 2223:22 \
  -e SSH_PASSWORD=your-secret \
  logsift-trae
```

## 本机 SSH 配置

把下面内容加入本机 SSH 配置文件。

Linux / macOS 常见位置：

```bash
~/.ssh/config
```

Windows OpenSSH 常见位置：

```text
C:\Users\<你的用户名>\.ssh\config
```

配置示例：

```sshconfig
Host logsift-trae
    HostName localhost
    User root
    Port 2222
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

## 在 Trae 中连接

在 Trae 中按 SSH 远程开发方式连接主机 `logsift-trae` 即可。

连接成功后，仓库路径为：

```text
/app
```

首次连接后，可在远端终端中执行：

```bash
cd /app
go test ./...
```

## 切换项目时的注意事项

- `localhost:2222` 同一时刻通常只能映射给一个容器
- 如果你想把 Trae 指到另一个项目容器，需要先停止当前占用 `2222` 的容器
- 然后让新容器重新映射 `2222:22`，再从 Trae 重连

例如：

```bash
docker stop logsift-trae
docker rm logsift-trae
```

## 故障排查

- 如果 Trae 卡在远端服务安装阶段，优先确认基础镜像不是 Alpine
- 如果 `docker build` 期间无法联网，优先检查代理参数
- 如果 SSH 能连但 Trae 服务启动失败，优先确认镜像是 Debian / Ubuntu / 其他 `glibc` 系
