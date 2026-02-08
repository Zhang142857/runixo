# TLS 证书管理

## 自动证书

Agent 首次启动时自动生成自签名 TLS 证书（有效期 10 年），包含服务器所有网络接口 IP 和 localhost。

## 重新生成

```bash
sudo rm -rf /var/lib/runixo/tls
sudo systemctl restart runixo-agent
```

重新生成后需在客户端重新导入证书。

## 客户端导入

**SSH 安装**：自动完成，无需手动操作。

**手动导入**：
1. 服务器上执行 `sudo cat /var/lib/runixo/tls/cert.pem`
2. 客户端 → 服务器操作菜单 →「导入证书」→ 粘贴

## 证书存储位置

| 平台 | 路径 |
|------|------|
| Windows | `%APPDATA%\Runixo\certificates\` |
| macOS | `~/Library/Application Support/Runixo/certificates/` |
| Linux | `~/.config/Runixo/certificates/` |

## 故障排查

**`ECONNRESET` 或 `certificate verify failed`**：客户端缺少服务器证书，手动导入或重新 SSH 安装即可。
