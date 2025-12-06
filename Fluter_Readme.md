# 饭盒售货机应用

## 项目概述

** 是一个高可靠、高客服用、产品级的自动饭盒售货机产品购买应用程序。该应用为用户提供了便捷的午餐购买体验，支持多种支付方式，并实现了设备管理和地理位置服务等功能。

## 项目特点

- **高可靠性**：稳定运行在各种环境下的售货机设备上
- **优质用户体验**：简洁直观的界面设计，流畅的操作流程
- **多平台支持**：支持Android和iOS平台
- **双支付渠道**：支持微信支付和支付宝支付
- **地理位置上报**：实时获取并上报设备位置信息
- **远程设备管理**：实现对设备的远程监控和管理

## 技术栈

- **前端框架**：Flutter
- **状态管理**：GetX (MVVM架构)
- **网络请求**：Dio
- **本地存储**：SharedPreferences
- **图片加载**：CachedNetworkImage
- **支付功能**：二维码生成 (qr_flutter)
- **动画效果**：Lottie
- **屏幕适配**：flutter_screenutil
- **地理位置**：geolocator
- **设备信息**：device_info_plus
- **日志记录**：logger

## 架构设计

### 架构模式

本项目采用 **MVVM (Model-View-ViewModel)** 架构模式，通过GetX框架实现：

- **Model**：数据模型层，负责数据的定义和存储
- **View**：视图层，负责UI展示和用户交互
- **ViewModel**：业务逻辑层，处理数据和UI之间的交互

### 项目结构

```
lib/
├── app/
│   ├── data/            # 数据层（模型、提供者、仓库）
│   ├── modules/         # 功能模块（各页面）
│   ├── routes/          # 路由管理
│   ├── services/        # 服务层（API、定位、支付等）
│   └── utils/           # 工具类（常量、扩展、助手等）
├── core/                # 核心功能
├── presentation/        # 展示层组件
└── main.dart            # 应用入口
```

## 核心功能

### 1. 设备管理

- 设备状态监控
- 设备信息展示
- 设备故障上报

### 2. 产品浏览与购买

- 产品分类展示
- 产品详情查看
- 购物车管理
- 订单创建与提交

### 3. 支付系统

- 微信支付
- 支付宝支付
- 支付状态管理
- 交易记录查询

### 4. 位置服务

- 设备地理位置上报
- 基于位置的服务

### 5. 用户中心

- 订单历史查询
- 个人信息管理
- 帮助中心

## 数据模型

### 数据结构层级

采用 "城市 > 设备 > 产品" 的层级结构：

1. **城市 (City)**：包含多个设备，是地理分区单位
2. **设备 (Device)**：位于特定城市，包含多个产品，是物理销售点
3. **产品 (Product)**：位于特定设备中，是可购买的商品

### 主要数据模型

- **用户 (User)**：用户基本信息和认证状态
- **订单 (Order)**：购买记录和状态
- **支付 (Payment)**：支付信息和交易状态
- **位置 (Location)**：地理坐标和相关信息

## 安装指南

### 环境要求

- Flutter 3.10.0 或更高版本
- Dart 3.0 或更高版本
- Android Studio / Xcode

### 安装步骤

1. 克隆项目
   ```bash
   git clone https://your-repository-url/e_lunchbox.git
   ```

2. 进入项目目录
   ```bash
   cd e_lunchbox
   ```

3. 安装依赖
   ```bash
   flutter pub get
   ```

## 运行指南

### 运行在Android设备上
```bash
flutter run --release --target-platform android-arm64
```

### 运行在iOS设备上
```bash
flutter run --release --no-codesign
```

### 构建APK
```bash
flutter build apk --release
```

### 构建IPA
```bash
flutter build ios --release
```

## 贡献指南

欢迎提交Issue和Pull Request来帮助改进这个项目！

1. Fork项目仓库
2. 创建你的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交你的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开一个Pull Request

## 许可证

本项目采用MIT许可证。详情请查看LICENSE文件。
