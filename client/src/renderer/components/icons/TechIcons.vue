<template>
  <img v-if="iconSrc" :src="iconSrc" :alt="name" class="tech-icon" @error="onError" />
  <svg v-else viewBox="0 0 32 32" fill="currentColor">
    <rect x="4" y="6" width="24" height="8" rx="2" opacity="0.8"/>
    <rect x="4" y="18" width="24" height="8" rx="2" opacity="0.8"/>
    <circle cx="8" cy="10" r="1.5" fill="#22c55e"/>
    <circle cx="8" cy="22" r="1.5" fill="#22c55e"/>
  </svg>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{ name: string }>()

const iconMap: Record<string, string> = {
  // 基础服务
  docker: 'apps/docker.png',
  nginx: 'apps/nginx.png',
  mysql: 'apps/mysql.png',
  redis: 'apps/redis.png',
  postgresql: 'apps/postgresql.png',
  mongodb: 'apps/mongodb.png',
  mariadb: 'apps/mariadb.png',
  mssql: 'apps/mssql.png',
  // 运行时 & 语言
  nodejs: 'apps/nodejs.png',
  python: 'apps/python.png',
  go: 'apps/go.png',
  java: 'apps/java.png',
  php: 'apps/php.png',
  dotnet: 'apps/dotnet.png',
  rust: 'apps/rust.png',
  ruby: 'apps/ruby.png',
  // 工具 & 管理
  git: 'apps/git.png',
  pm2: 'apps/pm2.png',
  process: 'apps/pm2.png',
  certbot: 'apps/ssl.png',
  ssl: 'apps/ssl.png',
  firewall: 'apps/firewall.png',
  backup: 'apps/backup.png',
  monitor: 'apps/monitor.png',
  // 云 & 平台
  cloudflare: 'apps/cloudflare.png',
  portainer: 'apps/portainer.png',
  // Web 服务器
  caddy: 'apps/caddy.png',
  tomcat: 'apps/tomcat.png',
  openresty: 'apps/openresty.png',
  openlitespeed: 'apps/openlitespeed.png',
  // CI/CD & DevOps
  jenkins: 'apps/jenkins.png',
  gitlab: 'apps/gitlab.png',
  gitea: 'apps/gitea.png',
  nexus: 'apps/nexus.png',
  sonarqube: 'apps/sonarqube.png',
  // 监控 & 可观测
  grafana: 'apps/grafana.png',
  prometheus: 'apps/prometheus.png',
  zabbix: 'apps/zabbix.png',
  'uptime-kuma': 'apps/uptime-kuma.png',
  influxdb: 'apps/influxdb.png',
  // 消息队列 & 中间件
  kafka: 'apps/kafka.png',
  rabbitmq: 'apps/rabbitmq.png',
  memcached: 'apps/memcached.png',
  consul: 'apps/consul.png',
  nacos: 'apps/nacos.png',
  // 应用
  wordpress: 'apps/wordpress.png',
  minecraft: 'apps/minecraft.png',
  minio: 'apps/minio.png',
  ollama: 'apps/ollama.png',
  n8n: 'apps/n8n.png',
  'code-server': 'apps/code-server.png',
  elasticsearch: 'apps/elasticsearch.png',
  clickhouse: 'apps/clickhouse.png',
  // 管理面板
  pgadmin: 'apps/pgadmin.png',
  phpmyadmin: 'apps/phpmyadmin.png',
  // 构建
  'static-build': 'apps/static.png',
  static: 'apps/static.png',
}

const loadError = ref(false)

const iconSrc = computed(() => {
  if (loadError.value) return null
  const filename = iconMap[props.name]
  if (filename) {
    return new URL(`../../assets/icons/${filename}`, import.meta.url).href
  }
  return null
})

function onError() {
  loadError.value = true
}
</script>

<style scoped>
.tech-icon {
  width: 65%;
  height: 65%;
  object-fit: contain;
  border-radius: 4px;
}

svg {
  width: 60%;
  height: 60%;
  opacity: 0.85;
}
</style>
