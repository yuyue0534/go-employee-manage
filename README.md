# emp-api — 员工管理系统接口服务

基于 Go + Gin + pgx，连接 Supabase PostgreSQL，无需认证，开箱即用。

---

## 快速开始

### 1. 配置环境变量

```bash
cp .env.example .env
```

`.env` 内容示例：

```env
DATABASE_URL=postgres://postgres:<password>@db.<ref>.supabase.co:5432/postgres
SERVER_PORT=8080
GIN_MODE=debug
```

> Supabase 连接串在控制台 **Settings → Database → Connection string → URI** 处获取。
> 建议使用 **Transaction Pooler**（端口 6543）以节省连接数。

### 2. 安装依赖并运行

```bash
go mod tidy
go run .
```

### 3. Docker 运行

```bash
docker build -t emp-api .
docker run -p 8080:8080 --env-file .env emp-api
```

---

## API 接口总览（共 19 个端点）

### 基础

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |

---

### 员工 `/api/v1/employees`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/employees` | 分页列表 |
| POST | `/api/v1/employees` | 新建员工 |
| GET | `/api/v1/employees/:id` | 员工详情（含当前部门、职位、薪资）|
| PATCH | `/api/v1/employees/:id` | 更新员工信息 |
| DELETE | `/api/v1/employees/:id` | 删除员工 |

**列表查询参数**：`page` / `size` / `gender`（M或F）/ `name`（模糊搜索）

**新建员工 Request Body**

```json
{
  "birth_date": "1990-05-20",
  "first_name": "Zhang",
  "last_name": "Wei",
  "gender": "M",
  "hire_date": "2024-01-10"
}
```

**员工详情 Response**

```json
{
  "code": 0, "message": "ok",
  "data": {
    "emp_no": 10001,
    "first_name": "Georgi", "last_name": "Facello",
    "gender": "M",
    "birth_date": "1953-09-02T00:00:00Z",
    "hire_date": "1986-06-26T00:00:00Z",
    "department": { "dept_no": "d005", "dept_name": "Development" },
    "current_title": "Senior Engineer",
    "current_salary": 88958
  }
}
```

---

### 薪资 `/api/v1/employees/:id/salaries`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/employees/:id/salaries` | 薪资历史（时间倒序）|
| POST | `/api/v1/employees/:id/salaries` | 新增薪资记录 |
| PATCH | `/api/v1/employees/:id/salaries/current` | 修改当前生效薪资金额 |
| DELETE | `/api/v1/employees/:id/salaries/:from_date` | 删除指定薪资记录 |

---

### 职位/头衔 `/api/v1/employees/:id/titles`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/employees/:id/titles` | 职位历史（时间倒序）|
| POST | `/api/v1/employees/:id/titles` | 分配新职位 |
| DELETE | `/api/v1/employees/:id/titles/:title/:from_date` | 删除指定职位记录 |

---

### 部门 `/api/v1/departments`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/departments` | 所有部门列表 |
| POST | `/api/v1/departments` | 新建部门 |
| GET | `/api/v1/departments/:id` | 部门详情 |
| PATCH | `/api/v1/departments/:id` | 修改部门名称 |
| DELETE | `/api/v1/departments/:id` | 删除部门（级联删除关联数据）|
| GET | `/api/v1/departments/:id/employees` | 部门当前在职员工 |
| GET | `/api/v1/departments/:id/manager` | 部门当前经理 |
| PUT | `/api/v1/departments/:id/manager` | 指派/更换部门经理 |

---

## 统一响应格式

```json
{ "code": 0, "message": "ok", "data": { ... } }
```

分页响应 `data` 包含：`items` / `total` / `page` / `size`

| code | 含义 |
|------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 404 | 资源不存在 |
| 409 | 冲突（如 dept_no 重复）|
| 500 | 服务器内部错误 |

---

## 业务约定

- `to_date = '9999-01-01'` 表示该记录**当前有效**（数据集历史设计约定）。
- 删除员工会因外键 `ON DELETE CASCADE` 自动级联删除薪资、职位、部门关联等所有记录。
- 员工详情同时返回当前部门、职位、薪资，均通过 `to_date = '9999-01-01'` 过滤。
