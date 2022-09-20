
drop table if exists zzfy;
create table zzfy (
    `id` int(11) not null auto_increment primary key,
    `source` varchar(20) comment '数据来源',
    `uid` varchar(50) default '' comment '用户 UID',
    `nickname` varchar(100) default '' comment '昵称',
    `username` varchar(100) default '' comment '用户名',
    `password` varchar(100) default '' comment '密码',
    `salt` varchar(20) default '' comment '盐值',
    `secques` varchar(50) default '' comment '安全码/交易密码',
    `mobile` varchar(100) default '' comment '手机号码',
    `email` varchar(100) default '' comment '邮箱',
    `qq` varchar(30) default '' comment 'QQ号',
    `realname` varchar(100) default '' comment '姓名',
    `gender` varchar(10) default '' comment '性别',
    `bday` char(10) default '' comment '生日',
    `idno` varchar(30) default '' comment '身份证号',
    `bankno` varchar(50) default '' comment '银行卡号',
    `address` varchar(255) default '' comment '地址',
    `note` varchar(255) default '' comment '备注/注释'
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
