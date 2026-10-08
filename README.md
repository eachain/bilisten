# Bilisten

Bilisten(Bilibili+Listen，必听)：一款基于bilibili的听歌应用。



**声明：**本代码库仅用于学习，不可用于任何商业活动！

## 编译

```
git clone git@github.com:eachain/bilisten.git
cd bilisten
wails3 package
open bin/Bilisten.app
```



## 截图

主页

![主页](./snapshots/index.png)



播放列表

![播放列表](snapshots/playlist.png)



搜索

![搜索](snapshots/search.png)



设置

![设置](snapshots/settings.png)



主题

![主题](snapshots/theme.png)

## 项目结构

- `backend/`: 后端代码
- `frontend/`: 前端代码
- `main.go`: 代码入口

## 用到的技术及三方库

1. [wails3](https://v3.wails.io/)：桌面应用包装
2. [vue3](https://cn.vuejs.org/)：响应式框架
3. [tailwindcss](https://www.tailwindcss.cn/)：布局及部分样式微调
4. [daisyui](https://daisyui.com/)：前端组件
5. [hugeicons](https://hugeicons.com/)：图标
6. [dashjs](https://dashjs.org/)：播放音频
7. [qrcode](https://github.com/soldair/node-qrcode)：生成登录二维码
8. [zod](https://zod.dev/)：数据校验
9. [vue-draggable-plus](https://vue-draggable-plus.pages.dev/)：设置音质选择
10. [vue-virtual-scroller](https://vue-virtual-scroller.netlify.app/)：优化播放列表滚动



## 再次声明

本代码库仅用于学习，不可用于任何商业活动！
