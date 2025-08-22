schema-manager init // 克隆到用户目录，缓存
schema-manager init -f // 删除缓存克隆
shcema-manager list // 列出 .opencmd/commands 下面所有的 .hl 文件 按照文件的目录树
schema-manager search xx // 搜索指定 .hl 是否存在，支持正则表达式, 这个搜索和文件夹无关，只搜文件部分
schema-manager check // 检查远程分支和现在分支是否一致，就看本地是否落后远程分支