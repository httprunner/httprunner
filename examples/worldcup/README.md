# World Cup Live

```text
$ wcl -h
Monitor FIFA World Cup Live

Usage:
  wcl [flags]

Flags:
      --android string      run android app
      --auto                auto enter live
  -d, --duration int        set duration in seconds (default 30)
  -h, --help                help for wcl
  -i, --interval int        set interval in seconds (default 15)
      --ios string          run ios app
  -l, --log-level string    set log level (default "INFO")
  -n, --match-name string   specify match name
  -p, --perf strings        specify performance monitor, e.g. sys_cpu,sys_mem,sys_net,sys_disk,fps,network,gpu
  -u, --uuid string         specify device serial or udid
  -v, --version             version for wcl
```


```bash
$ wcl -n "比利时vs摩洛哥" --android com.ss.android.ugc.aweme -d 300 -i 15 -u caf0cd51
$ wcl -n "比利时vs摩洛哥" --ios com.ss.iphone.ugc.Aweme -d 300 -i 15 -p sys_cpu,sys_mem,sys_disk,sys_net,fps,network,gpu -u 00008030-000438191421802E
```

## bundle id

| app | iOS | Android |
| -- | -- | -- |
| 抖音 | com.ss.iphone.ugc.Aweme | com.ss.android.ugc.aweme |
| 央视频 | com.cctv.yangshipin.app.iphone | com.cctv.yangshipin.app.androidp |
| 咪咕视频 | com.wondertek.hecmccmobile | com.cmcc.cmvideo |
