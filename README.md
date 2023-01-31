# M3U Merger

## 自定义

1. fork 本仓库

2. 开启 `Workflow` 读写权限

- 打开 https://github.com/你的名字/itv/settings/actions
- 将`Workflow permissions` 勾选 `Read and write permissions`

3. 新建一个 GitHub Token
   - 打开 https://github.com/settings/tokens
   - `Generate new token` - `Generate new token (classic)` 名字填 `GITHUB_TOKEN`

源列表 `source.txt`

关键字 `keywords.txt` (只会合并匹配上关键字的节目)

## 订阅地址

`https://fastly.jsdelivr.net/gh/popeyelau/itv@main/merged.m3u`
