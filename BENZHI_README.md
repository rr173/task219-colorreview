基于 Go 实现的纺织染色批次色差证据复核全栈 Web 项目，一款工艺分析服务，处理多点测色、色差归因与复核结论版本化发布。

# BENZHI 评测说明 — task219-colorreview

纺织染色批次色差证据复核台（全栈 Web 应用）。本文件说明评测构建、双架构 Docker 与 `--smoke-test` 契约。

## 构建命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
go run ./cmd/task219-colorreview --smoke-test
```

## Docker 双架构

```bash
bash build_benzhi_docker.sh my-project linux/amd64
bash build_benzhi_docker.sh my-project linux/arm64
```

镜像 ENTRYPOINT 为 `/app/colorreview`，CMD 为 `["--smoke-test"]`。运行冒烟测试时**只传 flag，不追加路径参数**：

```bash
docker run --rm my-project --smoke-test
```

## `--smoke-test` 契约

不启动长驻服务，而是端到端复现业务闭环并验证重启恢复：

1. 创建染色批次并推进到待复核；
2. 上传浴液温度曲线；
3. 记录仪器校准并上传 5 个测色点（含 1 个偏红异常点）；
4. 按 CIE2000 计算色差，断言异常点数量为 1；
5. 提交并确认工艺证据，剔除异常点；
6. 创建并发布复核结论；
7. 关闭数据库后重开，断言批次状态、测色点数量、结论状态与版本完整恢复。

全部通过后打印 `SMOKE TEST PASSED` 并以退出码 0 结束；任一断言失败返回非零退出码。

## 环境约定

- Go：1.26.3（`GOTOOLCHAIN=local`），`CGO_ENABLED=0`
- 依赖代理：`GOPROXY=https://goproxy.cn,direct`、`GOSUMDB=sum.golang.google.cn`
- SQLite：`modernc.org/sqlite v1.52.0`（内置 SQLite 3.46.1）
