# 配置

Agent 配置文件位于 `/etc/runixo/agent.yaml`。

## 完整配置

```yaml
server:
  host: "0.0.0.0"      # 监听地址
  port: 9527           # gRPC 端口
  tls:
    enabled: true      # 启用 TLS（强烈推荐）

auth:
  token: "your-token"  # 认证令牌

metrics:
  interval: 2          # 指标采集间隔（秒）

log:
  level: "info"        # debug, info, warn, error

update:
  auto: false          # 自动更新
  interval: 3600       # 检查间隔（秒）
```

修改后重启服务：`sudo systemctl restart runixo-agent`

## 服务管理

```bash
sudo systemctl status runixo-agent    # 查看状态
sudo systemctl start runixo-agent     # 启动
sudo systemctl stop runixo-agent      # 停止
sudo systemctl restart runixo-agent   # 重启
sudo journalctl -u runixo-agent -f    # 查看日志
```

## Token 管理

```bash
runixo-agent --gen-token              # 生成新 Token
```
