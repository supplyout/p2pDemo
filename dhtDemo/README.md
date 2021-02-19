# 这是用来测试dht的demo

## 测试内容：

1. 利用ipfs公共节点作为bootstrapPeer，实现多个节点的互相发现、连接功能
2. 在本地创建一个bootstrapPeer 节点，实现多个节点的互相发现、连接功能

## 文件介绍：

1. bootstrap.go 运行在本地的bootstrap节点
2. norm.go 普通节点，成功连接到某个bootstrap节点后，自己也可以成为一个bootstrap节点

## 用法：

1. 利用公共节点作为bootstrapPeer：

   ```cmd
   # 使用buildw.bat脚本构建
   buildw.bat
   # 首先运行节点创建一个room
   norm.exe -room myroom
   # 在另外一个终端再运行一个节点
   norm.exe -joinRoom myroom
   ```

2. 使用本地创建的bootstrapPeer

   ```cmd
   # 使用buildw.bat脚本构建
   buildw.bat
   # 首先运行bootstrap节点
   bootstrap.exe
   # 输出如下内容：
   # Addr:/ip4/10.0.0.193/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
   #Addr:/ip4/127.0.0.1/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
   # 这两个地址供其他norm节点，进行连接
   
   
   # 再运行一个节点并创建一个room
   norm.exe -room myroom -b /ip4/10.0.0.193/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
   # 在另外一个终端再运行一个节点
   norm.exe -joinRoom myroom -b /ip4/10.0.0.193/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
   
   ```

   ## 