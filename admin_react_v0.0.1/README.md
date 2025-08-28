# Trusioo Admin (admin_react_v0.0.1)

一个基于 React 18 + TypeScript + Vite 的前后端分离管理端项目脚手架，采用简化的 Feature-based 模块化目录结构，内置 React Router、Ant Design、TanStack Query 与 Zustand，可快速扩展与落地业务功能。

## 技术栈
- 构建工具：Vite 5 + @vitejs/plugin-react
- 前端框架：React 18 + TypeScript 5
- UI 组件库：Ant Design 5
- 路由：React Router v6
- 状态与数据：Zustand、@tanstack/react-query
- HTTP：Axios
- 代码规范：ESLint + Prettier

## 目录结构
```
admin_react_v0.0.1/
├─ public/
│  └─ index.html               # HTML 模板
├─ src/
│  ├─ app/                     # 应用级入口与全局的 Providers / Router
│  │  ├─ App.tsx
│  │  ├─ providers.tsx
│  │  └─ router.tsx
│  ├─ features/                # 功能模块（建议：用户、角色、权限、订单等）
│  │  └─ example/
│  │     ├─ components/
│  │     ├─ api/
│  │     ├─ hooks/
│  │     ├─ types.ts
│  │     └─ index.ts
│  ├─ shared/                  # 可复用的通用模块（UI、hooks、utils、api 基础等）
│  ├─ layouts/                 # 页面布局（侧边栏、顶部导航等）
│  │  └─ AdminLayout.tsx
│  ├─ assets/
│  │  └─ styles/
│  │     └─ index.css          # 全局样式
│  └─ config/                  # 配置（常量、环境、主题等）
├─ tsconfig.json
├─ tsconfig.node.json
├─ vite.config.ts
├─ package.json
└─ README.md
```

### 模块化约束（建议）
- feature 模块之间不要直接相互依赖，推荐通过 shared 层的抽象间接依赖。
- shared 不能依赖 features，保持单向依赖，避免循环。
- 每个模块尽量通过 `index.ts` 暴露其公共 API（组件、hooks、types）。
- 保持扁平目录，避免过深嵌套；按需新增子目录（components/api/hooks）。

## 快速开始
1) 安装依赖
```
npm install
```

2) 本地开发（默认端口 5173）
```
npm run dev
```
打开浏览器访问：http://localhost:5173/

3) 构建生产包
```
npm run build
```

4) 预览构建产物
```
npm run preview
```

## 推荐开发规范
- 组件文件使用大驼峰命名（如 UserTable.tsx），非组件文件使用小驼峰或中划线。
- 使用绝对路径别名（@app/@features/@shared/@layouts/@assets/@config），避免相对路径地狱。
- 使用 React Query 管理服务端数据，请求层统一封装在各模块的 api 目录中。
- 页面布局统一放在 layouts；登录页等特殊页面可直接在路由中挂载。
- 提交前运行 `npm run lint` 与 `npm run format`，保证代码风格统一。

## 如何新增功能模块（示例）
以新增“用户管理（user）”模块为例：
```
src/features/user/
├─ components/
│  ├─ UserTable.tsx
│  └─ UserForm.tsx
├─ api/
│  └─ userApi.ts               # axios 封装与 react-query hooks
├─ hooks/
│  └─ useUserStore.ts          # zustand store（可选）
├─ types.ts                    # 模型类型定义
└─ index.ts                    # 统一导出
```
在路由中按需引入页面组件，或在 AdminLayout 的菜单中增加入口。

## 环境变量
- 使用 Vite 的环境变量机制（.env、.env.development、.env.production）。
- 访问变量使用 `import.meta.env.VITE_*` 前缀。

## 常见问题
- 如果 eslint 安装报 peer 依赖冲突，请升级/降级 eslint 或插件版本保持兼容（本项目已对齐 eslint@^8 与插件）。
- Ant Design 在中文环境的字体与字号可根据视觉需求在 `assets/styles/index.css` 中调整。

---
如需我进一步搭建示例模块（用户/角色/权限等），或对接后端 API、登录鉴权、主题切换、国际化（i18n）等，请告诉我你的具体需求。