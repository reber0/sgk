# sgk
使用 Manticore Search 搭建社工库

### 安装

通过 [这里](https://manual.manticoresearch.com/Installation#Installation) 安装

### 使用流程
1. 通过 Manticore Search 的 indexer 生成数据库的全文索引(将索引与数据的 id 一一对应)
2. 通过 Manticore Search 的 searchd 启动一个服务
3. 通过向 searchd 服务发送关键字获取数据库中符合条件的索引 id
4. 通过索引 id 在数据库中获取相应数据

### 配置文件编写

涉及到两个地方的配置，一个是 manticore.conf，一个是 web/sgk_index_msg.json

* 现有数据库结构

    ```sql
    CREATE TABLE `3pk_com`.`users` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `uname` varchar(100) DEFAULT NULL COMMENT '用户名',
    `pass` varchar(100) DEFAULT NULL COMMENT '密码',
    `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
    `qqmail` varchar(100) DEFAULT NULL COMMENT 'QQ 邮箱',
    `mobile` varchar(100) default '' comment '手机号码',
    `address` varchar(255) DEFAULT NULL COMMENT '地址',
    PRIMARY KEY (`id`)
    ) ENGINE=InnoDB AUTO_INCREMENT=2000 DEFAULT CHARSET=utf8
    ```

    社工库涉及到的字段一般如下，我们对涉及到的所有字段统一别名：  
    数据来源，UID，昵称，用户名，密码，盐值，安全码/交易密码，手机号，邮箱，QQ 号，真实姓名，身份证号，银行卡号，地址，备注  
    source, uid, nickname, username, password, salt, secques, mobile, email, qq, realname, idno, bankno, address, note  

* manticore.conf 主要是查询的字段(需要建立索引的字段)

    上面数据库 3pk_com，表名为 users，我们约定配置文件中的 source 名字为 数据库名_表名，即 3pk_com_users  
    表中一般需要索引的字段为 uname、email、qqmail、mobile，密码、地址这种一般不做索引

    若单个字段有多列值，比如多个邮箱可以用 concat_ws 连接起来:  
    uname as username,concat_ws(", ", email, qqmail) as email,mobile

    编写 manticore.conf 添加

    ```
    source 3pk_com_users:base_source { # source 名字为 3pk_com_users
        sql_db = 3pk_com # 数据库名为 3pk_com
        sql_query = select id,uname as username,concat_ws(", ", email, qqmail)email,mobile from users
    }
    index 3pk_com_users:base_index { # index 名字也为 3pk_com_users
        source = 3pk_com_users # 这里是 source 的名字
        path = ./data/index/3pk_com_users # 存放索引的位置
    }
    ```

* sgk/web/sgk_index_msg.json

    查询时就需要显示所有字段了，在 web/sgk_index_msg.json 中添加如下配置(这里密码、地址这两列也查)

    ```json
    [
        {
            "index": "3pk_com_users",
            "db_name": "3pk_com",
            "table_name": "users",
            "columns": "id,uname as username,pass as password,concat_ws(', ', email, qqmail)email,mobile,address"
        }
    ]
    ```

### 创建索引
* 创建单个索引

    indexer -c ./manticore.conf 3pk_com_users

* 创建所有索引

    indexer -c ./manticore.conf --all

* searchd 运行中创建索引

    indexer -c ./manticore.conf 3pk_com_users --rotate

### 启动 searchd
searchd -c ./manticore.conf --console

### 搜索
* 正常搜索

    需要完全匹配单词

* 模糊搜索(* 匹配多个字符、?匹配单个字符)

    123*   得到 123xxx、123abc  
    \*aaa* 得到 111aaa、333aaa333、aaaccc  
    aa*cc  得到 aadcc、aabbcc、aa123cc  
    aa?cc  得到 aadcc、aa3cc  
    王?二  得到 王小二、王大二
