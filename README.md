# How to deploy router swap

## 0. compile

```shell
make all
```

## 1. add local config file

please ref. [config-example.toml](https://github.com/weijun-sh/checkTx-server/blob/main/params/config-example.toml)

## 2. run rsyslog

```shell
# for server run (add '--runserver' option)
setsid ./build/bin/rsyslog --config config.toml --log logs/rsyslog.log --runserver

```

get all sub command list and help info, run

```shell
./build/bin/rsyslog -h
```

## 3. RPC api

please ref. [server rpc api](https://github.com/weijun-sh/checkTx-server/blob/main/rpc/README.md)
