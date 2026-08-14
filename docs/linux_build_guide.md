# MaiGoDX Linux 版本编译与运行指南

如杂鱼大哥哥所愿，本次编译完全在你的本地电脑上通过 Go 跨平台交叉编译完成，未动用任何沙盒资源♡

## 1. 产物信息

- **文件名**：`maigodx-linux-amd64`
- **存放路径**：`D:\dev\MaiGoDX\maigodx-linux-amd64`
- **目标平台**：`Linux amd64` (`CGO_ENABLED=0` 纯 Go 静态编译，包含内嵌前端与 SQLite 纯 Go 驱动，开箱即用)

---

## 2. 在 Linux 服务器上运行指南

1. 将编译好的 `maigodx-linux-amd64` 上传至你的 Linux 服务器目标目录。
2. 赋予执行权限：
   ```bash
   chmod +x maigodx-linux-amd64
   ```
3. 启动服务（支持通过环境变量配置端口或 AimeDB 端口）：
   ```bash
   ./maigodx-linux-amd64
   ```
4. 若需后台常驻运行，可配合 `screen`、`tmux` 或编写 `systemd` 服务守护进程。
