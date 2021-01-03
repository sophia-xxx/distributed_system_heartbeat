# distributed_system_heartbeat

## 在VM上启动服务

启动VM，运行所有go文件，然后VM会运行一个Daemon进程，用来接受UDP message。


`$ go run *.go`+[-flags]

`-introducer`  这个flag标志当前VM/当前进程是introducer（介绍其他节点加入，即发送自己的membership list给刚加入的节点）

`-VM` 这个flag代表此程序运行在VM上（flag -host仅和此flag相关）

`-gossip` 这个flag代表心跳机制是 gossip 类型，默认为 all-to-all 类型（所有加入节点必须运行同一类型）

`-host` 这个flag定义VM的标号（01-10）

`-port` 这个flag定义程序的端口（本地运行时，多端口模拟多台VM）

例如，如果你想在虚拟机01上，以 gossip 心跳机制启动 introducer， 可以运行命令 `$ go run *.go -VM -introducer -host 01 -port 8002 -gossip`

如果你想在本地以all-to-all机制运行普通进程，可以使用命令 `$ go run *.go -host localhost -port 8002`


## 命令

* JOIN

`$ join`

离开的进程/节点或者是新加入的进程/节点， 运行命令 `$ join [introducerhost] [introducerport]`  来加入分布式系统。例如, `join 02 8001`

* LEAVE

`$ leave`

进程/节点会multicast LEAVE message，然后改变status="LEAVED"。离开的进程/节点将不再发送和接受心跳信息。

* FAIL

`Ctrl+C`

模拟突然宕机的进程/节点，它们将不再发送心跳。系统中的所有节点将标记他们的status为"FAILED"。

* Membership List

`display member`

列出所有member（ID+状态）

`display id`

列出当前进程/节点的ID

## All-to-All 心跳机制

此系统默认加入时为all-to-all心跳机制。如果在运行过程中想改变心跳机制，在当前节点运行 `$switch` 命令，然后所有节点会自动全部改变成另外一个类型


## Gossip 心跳机制
`$ -gossip`

可以通过启动时用-gossip flag 运行gossip心跳机制，也可以通过 `$switch` 命令改变类型。
