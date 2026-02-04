# 后端环境说明

## MySQL 连接配置
后端通过环境变量读取数据库连接信息，示例如下：

```
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=app
MYSQL_PASSWORD=password
MYSQL_DB=app_db
MYSQL_PARAMS=charset=utf8mb4&parseTime=True&loc=Local
```

参数含义：
- `MYSQL_HOST`：数据库 IP 或域名
- `MYSQL_PORT`：数据库端口
- `MYSQL_USER`：数据库用户名
- `MYSQL_PASSWORD`：数据库密码
- `MYSQL_DB`：数据库名
- `MYSQL_PARAMS`：连接参数（如字符集与时区等）

## 使用方式
在本地开发时可参考 `backend/.env.example`，将示例内容复制到你的环境变量配置中。

## JWT 配置
登录成功后会生成 JWT，用于访问其他 API。

```
JWT_SECRET=replace_with_long_random_string
JWT_ISSUER=sarah-project
JWT_TTL_MINUTES=60
```

参数含义：
- `JWT_SECRET`：签名密钥（请使用足够长的随机字符串）
- `JWT_ISSUER`：签发者标识
- `JWT_TTL_MINUTES`：Token 有效期（分钟）

## Customer API Key 存储
Customer 相关接口使用 API Key 与商户名双重验证，请在请求头中携带 `X-API-Key` 与 `X-Merchant-Name`。

API Key 和商户名不再放在环境变量中，而是存储在数据库 `customer_api_keys` 表中。
需要插入一条有效记录（`active = 1`）才能访问 customer 接口。
