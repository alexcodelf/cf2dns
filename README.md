# cf2dns

cf2dns 是一个基于github actions的优选 Cloudflare/GCore IP 并同步到 Cloudflare DNS 解析服务的工具。

## 功能特点

- 自动同步 Cloudflare 的 DNS 记录到其他 DNS 服务商
- 2小时自动更新 DNS 记录
- 支持自定义同步规则和过滤器


## 快速开始

在根目录下创建 `config.yaml` 文件，并添加以下内容:

```yaml
cloudflare:
  url: https://www.wetest.vip/api/cf2dns/get_cloudflare_ip
  domain: dev.local
  names: cf1,cf2,cf3
gcore:
  url: https://www.wetest.vip/api/cf2dns/get_gcore_ip
  domain: dev.local
  names: gcore1,gcore2,gcore3
cloudflareApiToken: yourtoken
maxDelay: 500
```

cloudflareAPI 令牌获取地址: https://dash.cloudflare.com/profile/api-tokens

在用户 API 令牌页面创建 API 令牌，并添加以下权限:

权限选择：区域 -> DNS -> 编辑

区域资源选择 `所有区域`

保存后复制 API 令牌，填入 `config.yaml` 文件中 `cloudflareApiToken` 字段

执行

```bash
go run ./cmd/main.go
```

### github actions 使用

默认情况下，会每12小时更新一次，在项目的 settings->secrets and variables -> actions 中添加以下变量:

- `CLOUDFLARE_URL`: Cloudflare API 的 URL
- `CLOUDFLARE_DOMAIN`: 需要解析优选 Cloudflare IP 的域名， 如 dev.yourdomain.com
- `CLOUDFLARE_NAMES`: 需要解析优选 Cloudflare 的子域名列表，以逗号分隔， 如 cf1,cf2,cf3
- `CLOUDFLARE_API_TOKEN`: Cloudflare API 的 token
- `GCORE_URL`: GCore API 的 URL
- `GCORE_DOMAIN`: 需要解析优选 GCore IP 的域名
- `GCORE_NAMES`: 需要解析优选 GCore IP 的子域名列表
- `MAX_DELAY`: 最大延迟时间



# 注意

cf2dns 优选IP仅针对网络可获取的节点进行优选解析，不提供任何CDN服务，严禁用于任何非法用途。