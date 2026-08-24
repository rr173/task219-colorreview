# task219-colorreview — 纺织染色批次色差证据复核台

纺织工艺师核对染色批次色差来源的 Web 服务：判断偏差来自浴液条件、取样位置还是仪器校准。支持批次生命周期、浴液温度/pH 曲线、布样多点测色、测色仪校准、CIE 色差计算（ΔE76/ΔE94/ΔE2000）、异常点剔除、工艺证据标注与复核结论版本化发布。

## 业务闭环

导入染色批次 → 上传浴液曲线 → 多点测色（仪器校准）→ 色差计算与容差判定 → 剔除污点/补充证据 → 发布复核结论（不可变版本）→ 封存批次。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...

# 端到端自检（Docker 判据）
go run ./cmd/task219-colorreview --smoke-test

# 启动 HTTP 服务
go run ./cmd/task219-colorreview --addr :8080 --db colorreview.db
```

## 核心 API（前缀 /api）

| 能力 | 入口 |
| --- | --- |
| 创建/查询批次 | `POST /api/batches` · `GET /api/batches/{id}` |
| 推进/封存批次 | `POST /api/batches/{id}/advance` · `POST /api/batches/{id}/seal` |
| 浴液曲线 | `POST /api/batches/{id}/bath-curves` · `GET /api/batches/{id}/bath-curves` |
| 测色点 | `POST /api/batches/{id}/measure-points` · `GET /api/batches/{id}/measure-points` |
| 剔除污点 | `POST /api/batches/{id}/measure-points/{mid}/reject` |
| 仪器校准 | `POST /api/instruments/calibrations` · `GET /api/instruments/calibrations` |
| 色差计算 | `POST /api/batches/{id}/color-diff` |
| 工艺证据 | `POST /api/batches/{id}/evidences` · `POST /api/batches/{id}/evidences/{eid}/confirm` |
| 复核结论 | `POST /api/batches/{id}/conclusion` · `POST /api/batches/{id}/conclusion/{cid}/publish` · `POST /api/batches/{id}/conclusion/{cid}/supersede` |
| 自检/演示 | `GET /api/self-check` · `POST /api/demo/import` |

## 技术栈

- Go 1.26.3，纯 Go SQLite 驱动 `modernc.org/sqlite v1.52.0`（SQLite 3.46.1，CGO 无关）
- 标准库 `net/http`（Go 1.22+ 路由），无第三方 Web 框架

## 项目结构

```
cmd/task219-colorreview/main.go   # 入口（--addr / --db / --smoke-test）
internal/model/                   # 实体与领域错误
internal/store/                   # SQLite 建表迁移与 CRUD
internal/batch/                   # 批次生命周期状态机
internal/sampling/                # 测色采样与仪器校准
internal/colorimetry/             # CIE Lab 转换与 ΔE 计算
internal/evidence/                # 工艺证据与冲突判定
internal/review/                  # 复核结论与版本化
internal/service/                 # 编排层
internal/httpapi/                 # HTTP 路由与处理
internal/webui/                   # 说明页面
```
