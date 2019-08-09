# 消息推送接口

该接口用于向当前`在线客户端`推送消息

## 接口信息

* URL: `/service/push`
* Method: POST
* Content-Type: application/x-www-form-urlencoded

## 接口参数

| 参数 | 类型 | 必传 | 描述 |
| --- | --- | --- | --- |
| app_key | str | 是 | 合作方 app key |
| token | str | 是 | 认证 token |
| mode | str | 是 | 客户端接入模式: im、push |
| message | str | 是 | 向客户端推送的消息 |

## Example

#### Request

```bash
curl -X POST \
  http://127.0.0.1:8080/service/push \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'token=token&app_key=app_key&mode=im&message=are%20you%20ok%20%3F'
```

#### Response

```json
{
  "info": "ok"
}
```
