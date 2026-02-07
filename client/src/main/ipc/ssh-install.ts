import { Client } from 'ssh2'
import { ipcMain, BrowserWindow } from 'electron'
import { readFileSync } from 'fs'
import { homedir } from 'os'
import { join } from 'path'

interface SshInstallParams {
  host: string
  sshPort: number
  username: string
  authType: 'password' | 'key'
  password?: string
  keyPath?: string
}

function sendLog(win: BrowserWindow, text: string, type: 'info' | 'success' | 'error' = 'info') {
  win.webContents.send('ssh:install:log', { text, type })
}

function sshExec(conn: Client, cmd: string): Promise<{ stdout: string; stderr: string; code: number }> {
  return new Promise((resolve, reject) => {
    conn.exec(cmd, (err, stream) => {
      if (err) return reject(err)
      let stdout = '', stderr = ''
      stream.on('data', (d: Buffer) => { stdout += d.toString() })
      stream.stderr.on('data', (d: Buffer) => { stderr += d.toString() })
      stream.on('close', (code: number) => resolve({ stdout: stdout.trim(), stderr: stderr.trim(), code: code || 0 }))
    })
  })
}

export function registerSshHandlers() {
  ipcMain.handle('ssh:installAgent', async (event, params: SshInstallParams) => {
    const win = BrowserWindow.fromWebContents(event.sender)!
    const conn = new Client()

    return new Promise<{ success: boolean; port: number; token: string; error?: string }>((resolve) => {
      const connectConfig: any = {
        host: params.host,
        port: params.sshPort || 22,
        username: params.username,
        readyTimeout: 15000,
      }

      if (params.authType === 'password') {
        connectConfig.password = params.password
      } else {
        const keyPath = (params.keyPath || '~/.ssh/id_rsa').replace('~', homedir())
        try { connectConfig.privateKey = readFileSync(keyPath) }
        catch { return resolve({ success: false, port: 0, token: '', error: `无法读取密钥: ${keyPath}` }) }
      }

      conn.on('error', (err) => {
        sendLog(win, `❌ SSH 连接失败: ${err.message}`, 'error')
        resolve({ success: false, port: 0, token: '', error: err.message })
      })

      conn.on('ready', async () => {
        try {
          sendLog(win, '✓ SSH 连接成功', 'success')

          // 检测系统架构
          sendLog(win, '检测系统架构...')
          const archRes = await sshExec(conn, 'uname -m')
          const arch = archRes.stdout.includes('aarch64') || archRes.stdout.includes('arm64') ? 'arm64' : 'amd64'
          const os = 'linux'
          sendLog(win, `系统: ${os}/${arch}`, 'success')

          // 检查是否已安装
          const existing = await sshExec(conn, 'which runixo-agent 2>/dev/null')
          if (existing.code === 0) {
            sendLog(win, '⚠ Agent 已安装，将更新到最新版本')
          }

          // 下载 agent (目前仅支持 linux/amd64)
          sendLog(win, '下载 Runixo Agent...')
          if (arch !== 'amd64') {
            sendLog(win, `❌ 暂不支持 ${arch} 架构`, 'error')
            conn.end()
            return resolve({ success: false, port: 0, token: '', error: `不支持的架构: ${arch}` })
          }
          const dlUrl = 'https://raw.githubusercontent.com/Zhang142857/runixo/main/agent/runixo-agent-linux'
          const dlCmd = `curl -fsSL -o /tmp/runixo-agent "${dlUrl}" 2>&1 || wget -q -O /tmp/runixo-agent "${dlUrl}" 2>&1`
          const dlRes = await sshExec(conn, dlCmd)
          if (dlRes.code !== 0) {
            sendLog(win, `❌ 下载失败: ${dlRes.stderr || dlRes.stdout}`, 'error')
            conn.end()
            return resolve({ success: false, port: 0, token: '', error: '下载 Agent 失败' })
          }
          sendLog(win, '✓ 下载完成', 'success')

          // 安装
          sendLog(win, '安装 Agent...')
          await sshExec(conn, 'chmod +x /tmp/runixo-agent')
          await sshExec(conn, 'sudo mv /tmp/runixo-agent /usr/local/bin/runixo-agent')
          await sshExec(conn, 'sudo mkdir -p /etc/runixo /var/lib/runixo')
          sendLog(win, '✓ 安装完成', 'success')

          // 生成 token
          sendLog(win, '生成认证令牌...')
          const tokenRes = await sshExec(conn, 'runixo-agent --gen-token 2>&1')
          sendLog(win, `token output: ${tokenRes.stdout}`)
          const token = tokenRes.stdout.match(/:\s*(.+)/)?.[1]?.trim() || tokenRes.stdout.trim()
          if (!token) {
            sendLog(win, '❌ 生成令牌失败', 'error')
            conn.end()
            return resolve({ success: false, port: 0, token: '', error: '生成令牌失败' })
          }
          sendLog(win, '✓ 令牌已生成', 'success')

          // 写配置文件
          const port = 9527
          const configYaml = `server:\n  host: "0.0.0.0"\n  port: ${port}\n  api_port: 9528\nauth:\n  token: "${token}"\ndata:\n  dir: "/var/lib/runixo"\nlog:\n  level: "info"`
          await sshExec(conn, `echo '${configYaml}' | sudo tee /etc/runixo/agent.yaml > /dev/null`)
          sendLog(win, '✓ 配置文件已写入', 'success')

          // 创建 systemd 服务
          sendLog(win, '配置系统服务...')
          const serviceUnit = `[Unit]\nDescription=Runixo Agent\nAfter=network.target\n\n[Service]\nType=simple\nExecStart=/usr/local/bin/runixo-agent --config /etc/runixo/agent.yaml\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=multi-user.target`
          await sshExec(conn, `echo '${serviceUnit}' | sudo tee /etc/systemd/system/runixo-agent.service > /dev/null`)
          await sshExec(conn, 'sudo systemctl daemon-reload')
          await sshExec(conn, 'sudo systemctl enable runixo-agent')
          await sshExec(conn, 'sudo systemctl restart runixo-agent')
          sendLog(win, '✓ 服务已启动', 'success')

          // 等待服务就绪
          sendLog(win, '等待 Agent 就绪...')
          await new Promise(r => setTimeout(r, 2000))
          const checkRes = await sshExec(conn, `curl -sf http://127.0.0.1:9528/health 2>/dev/null || echo "waiting"`)
          if (checkRes.stdout.includes('waiting')) {
            await new Promise(r => setTimeout(r, 3000))
          }
          sendLog(win, '✓ Agent 已就绪！', 'success')

          conn.end()
          resolve({ success: true, port, token })
        } catch (e: any) {
          sendLog(win, `❌ 安装出错: ${e.message}`, 'error')
          conn.end()
          resolve({ success: false, port: 0, token: '', error: e.message })
        }
      })

      sendLog(win, `连接 ${params.username}@${params.host}:${params.sshPort || 22}...`)
      conn.connect(connectConfig)
    })
  })
}
