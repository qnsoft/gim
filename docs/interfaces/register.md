# 合作方注册接口

该接口用于合作方注册

## 接口信息

* URL: `/register`
* Method: POST
* Content-Type: application/x-www-form-urlencoded

## 接口参数

| 参数 | 类型 | 必传 | 描述 |
| --- | --- | --- | --- |
| app_key | str | 是 | 合作方 app key 种子数据 |
| title | str | 是 | 如公司名称、简称 |
| reset | bool | 否 | 是否刷新 app secret |

## Example

#### Request

```bash
curl -X POST \
  http://127.0.0.1:8080/register \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'app_key=app_key&title=test&reset=false'
```

#### Response

```json
{
  "app_key": "app_key",
  "app_secret": "app_secret",
  "title": "title"
}
```
