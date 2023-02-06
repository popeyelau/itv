# M3U Merger

## 自定义

1. fork 本仓库

2. 开启 `Workflow` 读写权限

- 打开 https://github.com/你的名字/itv/settings/actions
- 将`Workflow permissions` 勾选 `Read and write permissions`

3. 新建一个 GitHub Token
   - 打开 https://github.com/settings/tokens
   - `Generate new token` - `Generate new token (classic)` 名字填 `GITHUB_TOKEN`

4. 修改配置
```
#sub.yaml
- group: '港澳台' 
  urls:
    - "https://telegram-feiyangdigital.v1.mk/gudou.m3u"
    - "https://ghproxy.com/https://raw.githubusercontent.com/fanmingming/live/main/tv/m3u/global.m3u"
  keywords: "discovery,viu,tvb,hbo,eleven,香港,台湾,中天,东森,東森,纬来,緯來,靖天,八大,台视,臺視,寰宇,寰宇,博斯,龙华,龍華,影迷,爱尔达,愛爾達,中天,年代,民视,民視,翡翠,凤凰,鳳凰,星卫,星衛,臺灣"
```

## 订阅地址

`https://raw.githubusercontent.com/popeyelau/itv/main/merged.m3u`
`https://ghproxy.com/https://raw.githubusercontent.com/popeyelau/itv/main/merged.m3u`
